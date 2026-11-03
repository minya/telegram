package telegram

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
)

const maxWebhookBody = 10 << 20

// Webhook is the http.Handler returned by WebhookHandler.
type Webhook struct {
	secretToken string
	handle      UpdateHandler
	onError     func(upd *Update, err error)
	inFlight    sync.WaitGroup
}

// WebhookHandler returns an http.Handler that receives updates sent by
// Telegram to a webhook registered with SetWebhook.
//
// If secretToken is non-empty, requests must carry it in the
// X-Telegram-Bot-Api-Secret-Token header (set it via SetWebhookParams.SecretToken).
//
// Each decoded update is acknowledged with 200 at once and handled in its
// own goroutine, so a slow handler (a search followed by several messages)
// does not hold Telegram's connection open, make it redeliver the update on
// timeout, or delay later updates. The handler's context carries the
// request's values but is not cancelled when Telegram disconnects. Because
// the 200 is sent first, a failing handler never causes a redelivery;
// handler errors, and panics, are passed to onError, which may be nil.
// Call Wait during shutdown to let handlers that are still running finish.
func WebhookHandler(secretToken string, handle UpdateHandler, onError func(upd *Update, err error)) *Webhook {
	return &Webhook{secretToken: secretToken, handle: handle, onError: onError}
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
	go h.run(context.WithoutCancel(r.Context()), upd)
	w.WriteHeader(http.StatusOK)
}

func (h *Webhook) run(ctx context.Context, upd *Update) {
	defer h.inFlight.Done()
	defer func() {
		if p := recover(); p != nil && h.onError != nil {
			h.onError(upd, fmt.Errorf("panic handling update %d: %v\n%s", upd.UpdateID, p, debug.Stack()))
		}
	}()
	if err := h.handle(ctx, upd); err != nil && h.onError != nil {
		h.onError(upd, err)
	}
}

// Wait blocks until every acknowledged update has been handled, or until
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
