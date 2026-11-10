package telegram

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"sync"
)

const maxWebhookBody = 10 << 20

// DefaultMaxConcurrent is how many updates a Webhook handles at once unless
// WithMaxConcurrent says otherwise; it matches Telegram's default
// max_connections.
const DefaultMaxConcurrent = 40

// Webhook is the http.Handler returned by WebhookHandler.
type Webhook struct {
	secretToken string
	handle      UpdateHandler
	onError     func(upd *Update, err error)
	inFlight    sync.WaitGroup

	key func(*Update) (int64, bool)
	sem chan struct{}

	mu sync.Mutex
	// queues holds, per key with a handler running, the updates waiting
	// for it, in update_id order.
	queues map[int64][]job
}

type job struct {
	ctx context.Context
	upd *Update
}

// WebhookOption configures a Webhook.
type WebhookOption func(*Webhook)

// WithMaxConcurrent caps how many updates are handled at once (default
// DefaultMaxConcurrent); n < 1 removes the cap. Updates over the cap wait
// for a free slot, so a backlog Telegram replays after downtime does not
// start every handler at once.
func WithMaxConcurrent(n int) WebhookOption {
	return func(h *Webhook) {
		if n < 1 {
			h.sem = nil
		} else {
			h.sem = make(chan struct{}, n)
		}
	}
}

// WithOrderKey sets which updates are handled one after another: those
// with the same key run in update_id order, never two at once. The default
// is UpdateChatID, so a chat's messages and button presses are handled in
// order. A key function returning false lets an update run independently;
// nil makes every update independent.
func WithOrderKey(key func(*Update) (int64, bool)) WebhookOption {
	return func(h *Webhook) { h.key = key }
}

// UpdateChatID returns the chat an update belongs to: the chat of its
// message, edited message, channel post or the callback query's message,
// else the callback or inline query's sender, else the chat of a member
// update.
func UpdateChatID(upd *Update) (int64, bool) {
	for _, m := range []*Message{upd.Message, upd.EditedMessage, upd.ChannelPost, upd.EditedChannelPost} {
		if m != nil {
			return m.Chat.ID, true
		}
	}
	switch {
	case upd.CallbackQuery != nil && upd.CallbackQuery.Message != nil:
		return upd.CallbackQuery.Message.Chat.ID, true
	case upd.CallbackQuery != nil && upd.CallbackQuery.From != nil:
		return upd.CallbackQuery.From.ID, true
	case upd.InlineQuery != nil && upd.InlineQuery.From != nil:
		return upd.InlineQuery.From.ID, true
	case upd.MyChatMember != nil:
		return upd.MyChatMember.Chat.ID, true
	case upd.ChatMember != nil:
		return upd.ChatMember.Chat.ID, true
	}
	return 0, false
}

// WebhookHandler returns an http.Handler that receives updates sent by
// Telegram to a webhook registered with SetWebhook.
//
// If secretToken is non-empty, requests must carry it in the
// X-Telegram-Bot-Api-Secret-Token header (set it via SetWebhookParams.SecretToken).
//
// Each decoded update is acknowledged with 200 at once and handled in the
// background, so a slow handler (a search followed by several messages)
// does not hold Telegram's connection open, make it redeliver the update on
// timeout, or delay other chats' updates. Updates of one chat are handled
// one after another, in update_id order (WithOrderKey), and at most
// DefaultMaxConcurrent run at once (WithMaxConcurrent). The handler's
// context carries the request's values but is not cancelled when Telegram
// disconnects. Because the 200 is sent first, a failing handler never
// causes a redelivery; handler errors, and panics, are passed to onError,
// which may be nil. Call Wait during shutdown to let handlers that are
// still running or queued finish.
func WebhookHandler(secretToken string, handle UpdateHandler, onError func(upd *Update, err error), opts ...WebhookOption) *Webhook {
	h := &Webhook{
		secretToken: secretToken,
		handle:      handle,
		onError:     onError,
		key:         UpdateChatID,
		sem:         make(chan struct{}, DefaultMaxConcurrent),
		queues:      map[int64][]job{},
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// ServeHTTP implements http.Handler.
func (h *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.secretToken != "" {
		got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
		if subtle.ConstantTimeCompare([]byte(got), []byte(h.secretToken)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	upd := new(Update)
	body := http.MaxBytesReader(w, r.Body, maxWebhookBody)
	if err := json.NewDecoder(body).Decode(upd); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Counted before the 200 goes out, so a Wait that starts after the
	// server stopped accepting requests sees every acknowledged update.
	h.inFlight.Add(1)
	h.schedule(job{ctx: context.WithoutCancel(r.Context()), upd: upd})
	w.WriteHeader(http.StatusOK)
}

// schedule runs j now, or queues it behind the running update of its key.
func (h *Webhook) schedule(j job) {
	if h.key == nil {
		go h.run(j)
		return
	}
	key, ok := h.key(j.upd)
	if !ok {
		go h.run(j)
		return
	}
	h.mu.Lock()
	if q, running := h.queues[key]; running {
		i, _ := slices.BinarySearchFunc(q, j.upd.UpdateID, func(e job, id int64) int {
			switch {
			case e.upd.UpdateID < id:
				return -1
			case e.upd.UpdateID > id:
				return 1
			}
			return 0
		})
		h.queues[key] = slices.Insert(q, i, j)
		h.mu.Unlock()
		return
	}
	h.queues[key] = nil
	h.mu.Unlock()
	go h.runKey(key, j)
}

// runKey handles j and then the updates queued behind it for key.
func (h *Webhook) runKey(key int64, j job) {
	for {
		h.run(j)
		h.mu.Lock()
		q := h.queues[key]
		if len(q) == 0 {
			delete(h.queues, key)
			h.mu.Unlock()
			return
		}
		j, h.queues[key] = q[0], q[1:]
		h.mu.Unlock()
	}
}

func (h *Webhook) run(j job) {
	defer h.inFlight.Done()
	if h.sem != nil {
		h.sem <- struct{}{}
		defer func() { <-h.sem }()
	}
	ctx, upd := j.ctx, j.upd
	defer func() {
		if p := recover(); p != nil && h.onError != nil {
			h.onError(upd, fmt.Errorf("panic handling update %d: %v\n%s", upd.UpdateID, p, debug.Stack()))
		}
	}()
	if err := h.handle(ctx, upd); err != nil && h.onError != nil {
		h.onError(upd, err)
	}
}

// Wait blocks until every acknowledged update, queued ones included, has
// been handled, or until
// ctx is done, in which case it returns ctx's error.
func (h *Webhook) Wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() { h.inFlight.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
