package telegram

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Error is returned when the Bot API answers a request with "ok": false.
// Use errors.As to inspect it, e.g. to honour RetryAfter on HTTP 429.
type Error struct {
	Method          string
	Code            int
	Description     string
	RetryAfter      int   // seconds to wait before retrying (flood control)
	MigrateToChatID int64 // set when a group was migrated to a supergroup
}

func (e *Error) Error() string {
	return fmt.Sprintf("telegram: %s: %d %s", e.Method, e.Code, e.Description)
}

// ResponseParameters describes why a request was unsuccessful.
type ResponseParameters struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	RetryAfter      int   `json:"retry_after,omitempty"`
}

// redactedError hides the bot token in an error message while keeping the
// underlying cause available to errors.Is / errors.As.
type redactedError struct {
	msg string
	err error
}

func (e *redactedError) Error() string { return e.msg }
func (e *redactedError) Unwrap() error { return e.err }

// redact removes the bot token from err. Transport errors from net/http embed
// the request URL, which contains the token.
func (api *API) redact(err error) error {
	if err == nil || api.token == "" || !strings.Contains(err.Error(), api.token) {
		return err
	}
	cause := err
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		cause = urlErr.Err
	}
	return &redactedError{
		msg: strings.ReplaceAll(err.Error(), api.token, "<redacted>"),
		err: cause,
	}
}
