package telegram

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
