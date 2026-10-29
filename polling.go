package telegram

import (
	"context"
	"errors"
	"time"
)

// UpdateHandler processes a single update.
type UpdateHandler func(ctx context.Context, upd *Update) error

// PollOptions configures Poll. The zero value is ready to use.
type PollOptions struct {
	// Offset is the first update ID to fetch.
	Offset int64
	// Timeout is the long-poll timeout; defaults to 50s. It must be shorter
	// than the HTTP client's timeout.
	Timeout time.Duration
	// AllowedUpdates limits the update types received, e.g. []string{"message"}.
	AllowedUpdates []string
	// OnError is called when getUpdates fails (upd is nil) or when the
	// handler returns an error. Optional.
	OnError func(upd *Update, err error)
}

const (
	defaultPollTimeout = 50 * time.Second
	maxRetryDelay      = 30 * time.Second
)

// minRetryDelay is a variable so tests can shorten it.
var minRetryDelay = time.Second

// Poll receives updates with long polling and passes each to handle until ctx
// is cancelled, then returns ctx.Err(). Failed getUpdates calls are retried
// with exponential backoff, honouring Telegram's retry_after.
func Poll(ctx context.Context, api *API, handle UpdateHandler, opts PollOptions) error {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultPollTimeout
	}
	params := GetUpdatesParams{
		Offset:         opts.Offset,
		Timeout:        int(timeout / time.Second),
		AllowedUpdates: opts.AllowedUpdates,
	}

	var delay time.Duration
	for {
		updates, err := api.GetUpdates(ctx, params)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if opts.OnError != nil {
				opts.OnError(nil, err)
			}
			delay = retryDelay(delay, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		delay = 0

		for i := range updates {
			upd := &updates[i]
			if err := handle(ctx, upd); err != nil && opts.OnError != nil {
				opts.OnError(upd, err)
			}
			params.Offset = upd.UpdateID + 1
		}
	}
}

func retryDelay(prev time.Duration, err error) time.Duration {
	var apiErr *Error
	if errors.As(err, &apiErr) && apiErr.RetryAfter > 0 {
		return time.Duration(apiErr.RetryAfter) * time.Second
	}
	if prev < minRetryDelay {
		return minRetryDelay
	}
	return min(prev*2, maxRetryDelay)
}
