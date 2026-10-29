package telegram

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

const maxWebhookBody = 10 << 20

// WebhookHandler returns an http.Handler that receives updates sent by
// Telegram to a webhook registered with SetWebhook.
//
// If secretToken is non-empty, requests must carry it in the
// X-Telegram-Bot-Api-Secret-Token header (set it via SetWebhookParams.SecretToken).
// The handler always answers 200 once an update is decoded, so a failing
// handler does not make Telegram redeliver the update; handler errors are
// passed to onError, which may be nil.
func WebhookHandler(secretToken string, handle UpdateHandler, onError func(upd *Update, err error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if secretToken != "" {
			got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
			if subtle.ConstantTimeCompare([]byte(got), []byte(secretToken)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}

		var upd Update
		body := http.MaxBytesReader(w, r.Body, maxWebhookBody)
		if err := json.NewDecoder(body).Decode(&upd); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := handle(r.Context(), &upd); err != nil && onError != nil {
			onError(&upd, err)
		}
		w.WriteHeader(http.StatusOK)
	})
}
