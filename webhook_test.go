package telegram

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWebhookHandler(t *testing.T) {
	var got *Update
	var handlerErr error
	h := WebhookHandler("s3cret",
		func(ctx context.Context, upd *Update) error {
			got = upd
			return errors.New("handler failed")
		},
		func(upd *Update, err error) { handlerErr = err },
	)

	send := func(method, secret, body string) int {
		req := httptest.NewRequest(method, "/hook", strings.NewReader(body))
		if secret != "" {
			req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secret)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := send(http.MethodGet, "s3cret", ""); code != http.StatusMethodNotAllowed {
		t.Errorf("GET: %d", code)
	}
	if code := send(http.MethodPost, "", `{"update_id":1}`); code != http.StatusUnauthorized {
		t.Errorf("missing secret: %d", code)
	}
	if code := send(http.MethodPost, "wrong", `{"update_id":1}`); code != http.StatusUnauthorized {
		t.Errorf("wrong secret: %d", code)
	}
	if code := send(http.MethodPost, "s3cret", `not json`); code != http.StatusBadRequest {
		t.Errorf("bad body: %d", code)
	}
	if got != nil {
		t.Fatal("handler called for rejected request")
	}

	if code := send(http.MethodPost, "s3cret", `{"update_id":42}`); code != http.StatusOK {
		t.Errorf("valid update: %d", code)
	}
	if err := h.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got == nil || got.UpdateID != 42 {
		t.Errorf("handler got %+v", got)
	}
	if handlerErr == nil {
		t.Error("onError not called")
	}
}

// F68: the update is acknowledged before the handler runs, the handler's
// context survives Telegram disconnecting, and Wait covers handlers still
// running.
func TestWebhookAcknowledgesFirst(t *testing.T) {
	release := make(chan struct{})
	ctxErr := make(chan error, 1)
	h := WebhookHandler("", func(ctx context.Context, upd *Update) error {
		<-release
		ctxErr <- ctx.Err()
		return nil
	}, nil)

	reqCtx, disconnect := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"update_id":7}`)).WithContext(reqCtx)
	rec := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rec, req) // must return while the handler is still blocked
	if rec.Code != http.StatusOK || time.Since(start) > time.Second {
		t.Fatalf("status %d after %s, want 200 at once", rec.Code, time.Since(start))
	}
	disconnect() // Telegram drops the connection

	short, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := h.Wait(short); err == nil {
		t.Fatal("Wait returned while the handler was still running")
	}
	close(release)
	if err := <-ctxErr; err != nil {
		t.Fatalf("handler context = %v after the disconnect, want it alive", err)
	}
	if err := h.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookReportsPanics(t *testing.T) {
	reported := make(chan error, 1)
	h := WebhookHandler("", func(context.Context, *Update) error { panic("boom") },
		func(_ *Update, err error) { reported <- err })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"update_id":1}`)))
	h.Wait(context.Background())
	if err := <-reported; err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("onError got %v, want the panic", err)
	}
}

func post(h *Webhook, body string) {
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(body)))
}

// F116: one chat's updates run one after another in update_id order, even
// when they arrive out of order, while another chat's run alongside; Wait
// covers queued updates.
func TestWebhookOrdersUpdatesPerChat(t *testing.T) {
	var mu sync.Mutex
	var order []int64
	running := map[int64]int{}
	overlap := false
	firstIn := make(chan struct{})
	release := make(chan struct{})
	otherDone := make(chan struct{})
	h := WebhookHandler("", func(ctx context.Context, upd *Update) error {
		chat, _ := UpdateChatID(upd)
		mu.Lock()
		running[chat]++
		if running[chat] > 1 {
			overlap = true
		}
		order = append(order, upd.UpdateID)
		mu.Unlock()
		switch upd.UpdateID {
		case 1:
			close(firstIn)
			<-release // holds chat 5 until chat 9 has been handled
		case 100:
			close(otherDone)
		}
		mu.Lock()
		running[chat]--
		mu.Unlock()
		return nil
	}, nil)

	post(h, `{"update_id":1,"message":{"message_id":1,"chat":{"id":5}}}`)
	<-firstIn
	for _, id := range []string{"4", "2", "3"} { // arrive out of order
		post(h, `{"update_id":`+id+`,"message":{"message_id":1,"chat":{"id":5}}}`)
	}
	post(h, `{"update_id":100,"message":{"message_id":1,"chat":{"id":9}}}`)
	select {
	case <-otherDone:
	case <-time.After(3 * time.Second):
		t.Fatal("another chat's update waited for chat 5")
	}
	short, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if h.Wait(short) == nil {
		t.Fatal("Wait returned with updates still queued")
	}
	close(release)
	if err := h.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	var chat5 []int64
	for _, id := range order {
		if id < 100 {
			chat5 = append(chat5, id)
		}
	}
	if overlap || !slices.Equal(chat5, []int64{1, 2, 3, 4}) {
		t.Fatalf("chat 5 handled %v (overlap %v), want 1 2 3 4 one at a time", chat5, overlap)
	}
}

// F116: no more than the cap run at once.
func TestWebhookCapsConcurrency(t *testing.T) {
	var active, peak atomic.Int32
	h := WebhookHandler("", func(context.Context, *Update) error {
		n := active.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		active.Add(-1)
		return nil
	}, nil, WithMaxConcurrent(3))
	for i := range 30 {
		post(h, fmt.Sprintf(`{"update_id":%d,"message":{"message_id":1,"chat":{"id":%d}}}`, i+1, i+1))
	}
	if err := h.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if peak.Load() > 3 || peak.Load() < 2 {
		t.Fatalf("peak concurrent handlers = %d, want at most 3 (and some overlap)", peak.Load())
	}
}
