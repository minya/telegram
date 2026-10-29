package telegram

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	if got == nil || got.UpdateID != 42 {
		t.Errorf("handler got %+v", got)
	}
	if handlerErr == nil {
		t.Error("onError not called")
	}
}
