package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const testToken = "123456:SECRET-token"

// fakeServer serves Bot API methods from per-method handlers and records the
// last request each method received.
type fakeServer struct {
	t        *testing.T
	mu       sync.Mutex
	handlers map[string]http.HandlerFunc
	srv      *httptest.Server
}

func newFakeServer(t *testing.T) *fakeServer {
	fs := &fakeServer{t: t, handlers: map[string]http.HandlerFunc{}}
	fs.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := "/bot" + testToken + "/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			http.NotFound(w, r)
			return
		}
		method := strings.TrimPrefix(r.URL.Path, prefix)
		fs.mu.Lock()
		h, ok := fs.handlers[method]
		fs.mu.Unlock()
		if !ok {
			t.Errorf("unexpected method %q", method)
			http.NotFound(w, r)
			return
		}
		h(w, r)
	}))
	t.Cleanup(fs.srv.Close)
	return fs
}

func (fs *fakeServer) handle(method string, h http.HandlerFunc) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.handlers[method] = h
}

func (fs *fakeServer) api() *API {
	return NewAPI(testToken, WithBaseURL(fs.srv.URL))
}

func reply(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, body)
}

func TestSendMessageReturnsMessage(t *testing.T) {
	fs := newFakeServer(t)
	fs.handle("sendMessage", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if got["chat_id"] != float64(42) || got["text"] != "hi" {
			t.Errorf("unexpected params %v", got)
		}
		if _, ok := got["reply_markup"]; ok {
			t.Errorf("reply_markup should be omitted when nil")
		}
		reply(w, `{"ok":true,"result":{"message_id":7,"date":1,"chat":{"id":42,"type":"private"},"text":"hi"}}`)
	})

	msg, err := fs.api().SendMessage(context.Background(), SendMessageParams{ChatID: 42, Text: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if msg.MessageID != 7 || msg.Chat.ID != 42 {
		t.Errorf("unexpected message %+v", msg)
	}
}

func TestAPIErrorIsReturned(t *testing.T) {
	fs := newFakeServer(t)
	for _, method := range []string{"answerCallbackQuery", "setChatMenuButton", "setWebhook", "editMessageText"} {
		fs.handle(method, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			reply(w, `{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":5}}`)
		})
	}
	api := fs.api()
	ctx := context.Background()
	_, editErr := api.EditMessageText(ctx, EditMessageTextParams{ChatID: 1, MessageID: 2, Text: "x"})

	errs := map[string]error{
		"answerCallbackQuery": api.AnswerCallbackQuery(ctx, AnswerCallbackQueryParams{CallbackQueryID: "q"}),
		"setChatMenuButton":   api.SetChatMenuButton(ctx, SetChatMenuButtonParams{}),
		"setWebhook":          api.SetWebhook(ctx, SetWebhookParams{URL: "https://example.com"}),
		"editMessageText":     editErr,
	}
	for method, err := range errs {
		var apiErr *Error
		if !errors.As(err, &apiErr) {
			t.Errorf("%s: want *Error, got %v", method, err)
			continue
		}
		if apiErr.Method != method || apiErr.Code != 429 || apiErr.RetryAfter != 5 || apiErr.Description != "Too Many Requests" {
			t.Errorf("%s: unexpected error %+v", method, apiErr)
		}
	}
}

func TestTransportErrorRedactsToken(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close() // nothing listens here any more

	api := NewAPI(testToken, WithBaseURL("http://"+addr))
	_, err = api.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), testToken) {
		t.Errorf("error leaks token: %v", err)
	}
	if !strings.Contains(err.Error(), "<redacted>") {
		t.Errorf("expected redacted URL in error: %v", err)
	}

	_, err = api.DownloadFile(context.Background(), &File{FilePath: "documents/a.txt"})
	if err == nil || strings.Contains(err.Error(), testToken) {
		t.Errorf("download error leaks token or is nil: %v", err)
	}
}

func TestCancelledContextIsDetectable(t *testing.T) {
	fs := newFakeServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := fs.api().GetMe(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
}

func TestSendDocumentMultipart(t *testing.T) {
	fs := newFakeServer(t)
	fs.handle("sendDocument", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("chat_id"); got != "42" {
			t.Errorf("chat_id = %q", got)
		}
		if _, ok := r.MultipartForm.Value["caption"]; ok {
			t.Errorf("empty caption should not be sent")
		}
		var markup InlineKeyboardMarkup
		if err := json.Unmarshal([]byte(r.FormValue("reply_markup")), &markup); err != nil {
			t.Errorf("reply_markup is not JSON: %v", err)
		} else if markup.InlineKeyboard[0][0].CallbackData != "ok" {
			t.Errorf("unexpected markup %+v", markup)
		}
		f, hdr, err := r.FormFile("document")
		if err != nil {
			t.Fatal(err)
		}
		content, _ := io.ReadAll(f)
		if hdr.Filename != "a.txt" || string(content) != "hello" {
			t.Errorf("unexpected file %q %q", hdr.Filename, content)
		}
		reply(w, `{"ok":true,"result":{"message_id":9,"date":1,"chat":{"id":42,"type":"private"}}}`)
	})

	msg, err := fs.api().SendDocument(context.Background(), SendDocumentParams{
		ChatID:   42,
		Document: InputFile{FileName: "a.txt", Content: []byte("hello")},
		ReplyMarkup: InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{{
			{Text: "OK", CallbackData: "ok"},
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.MessageID != 9 {
		t.Errorf("unexpected message %+v", msg)
	}
}

func TestSendPhotoByRef(t *testing.T) {
	fs := newFakeServer(t)
	fs.handle("sendPhoto", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("photo"); got != "file-123" {
			t.Errorf("photo = %q", got)
		}
		reply(w, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":1,"type":"private"}}}`)
	})
	if _, err := fs.api().SendPhoto(context.Background(), SendPhotoParams{ChatID: 1, Photo: InputFile{Ref: "file-123"}}); err != nil {
		t.Fatal(err)
	}

	_, err := fs.api().SendPhoto(context.Background(), SendPhotoParams{ChatID: 1})
	if err == nil {
		t.Error("expected error for empty InputFile")
	}
}

func TestSetWebhookUploadsCertificate(t *testing.T) {
	fs := newFakeServer(t)
	fs.handle("setWebhook", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("url"); got != "https://example.com/hook" {
			t.Errorf("url = %q", got)
		}
		if got := r.FormValue("allowed_updates"); got != `["message"]` {
			t.Errorf("allowed_updates = %q", got)
		}
		if _, _, err := r.FormFile("certificate"); err != nil {
			t.Errorf("certificate missing: %v", err)
		}
		reply(w, `{"ok":true,"result":true}`)
	})
	err := fs.api().SetWebhook(context.Background(), SetWebhookParams{
		URL:            "https://example.com/hook",
		Certificate:    &InputFile{FileName: "cert.pem", Content: []byte("PEM")},
		AllowedUpdates: []string{"message"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEditMessageTextInlineReturnsNil(t *testing.T) {
	fs := newFakeServer(t)
	fs.handle("editMessageText", func(w http.ResponseWriter, r *http.Request) {
		reply(w, `{"ok":true,"result":true}`)
	})
	msg, err := fs.api().EditMessageText(context.Background(), EditMessageTextParams{InlineMessageID: "abc", Text: "x"})
	if err != nil || msg != nil {
		t.Errorf("want nil, nil; got %v, %v", msg, err)
	}
}

func TestDownloadFile(t *testing.T) {
	fs := newFakeServer(t)
	api := fs.api()

	if _, err := api.DownloadFile(context.Background(), &File{FileID: "big"}); err == nil {
		t.Error("expected error for empty FilePath")
	}

	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/file/bot"+testToken+"/documents/a.txt" {
			http.NotFound(w, r)
			return
		}
		io.WriteString(w, "content")
	}))
	defer fileSrv.Close()
	api = NewAPI(testToken, WithBaseURL(fileSrv.URL))

	content, err := api.DownloadFile(context.Background(), &File{FilePath: "documents/a.txt"})
	if err != nil || string(content) != "content" {
		t.Errorf("got %q, %v", content, err)
	}
	if _, err := api.DownloadFile(context.Background(), &File{FilePath: "missing"}); err == nil {
		t.Error("expected error for 404")
	}
}

func TestPollRetriesAndStops(t *testing.T) {
	minRetryDelay = 10 * time.Millisecond
	t.Cleanup(func() { minRetryDelay = time.Second })

	fs := newFakeServer(t)
	var mu sync.Mutex
	calls := 0
	var offsets []int64
	fs.handle("getUpdates", func(w http.ResponseWriter, r *http.Request) {
		var params GetUpdatesParams
		json.NewDecoder(r.Body).Decode(&params)
		mu.Lock()
		calls++
		n := calls
		offsets = append(offsets, params.Offset)
		mu.Unlock()
		switch n {
		case 1:
			w.WriteHeader(http.StatusBadGateway)
			io.WriteString(w, "<html>bad gateway</html>")
		case 2:
			reply(w, `{"ok":true,"result":[{"update_id":10,"message":{"message_id":1,"date":1,"chat":{"id":1,"type":"private"},"text":"a"}},{"update_id":11}]}`)
		default:
			reply(w, `{"ok":true,"result":[]}`)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	var handled []int64
	var pollErrs, handlerErrs int
	handler := func(ctx context.Context, upd *Update) error {
		handled = append(handled, upd.UpdateID)
		if upd.UpdateID == 11 {
			return errors.New("boom")
		}
		return nil
	}
	opts := PollOptions{
		Offset: 5,
		OnError: func(upd *Update, err error) {
			if upd == nil {
				pollErrs++
			} else {
				handlerErrs++
			}
		},
	}

	done := make(chan error, 1)
	go func() { done <- Poll(ctx, fs.api(), handler, opts) }()

	deadline := time.After(5 * time.Second)
	for {
		mu.Lock()
		n := calls
		mu.Unlock()
		if n >= 3 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("poll did not make progress")
		case <-time.After(5 * time.Millisecond):
		}
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Errorf("Poll returned %v", err)
	}

	if pollErrs != 1 || handlerErrs != 1 {
		t.Errorf("pollErrs=%d handlerErrs=%d", pollErrs, handlerErrs)
	}
	if len(handled) != 2 || handled[0] != 10 || handled[1] != 11 {
		t.Errorf("handled %v", handled)
	}
	mu.Lock()
	defer mu.Unlock()
	if offsets[0] != 5 || offsets[1] != 5 || offsets[2] != 12 {
		t.Errorf("offsets %v", offsets)
	}
}

func TestRetryDelay(t *testing.T) {
	if d := retryDelay(0, errors.New("x")); d != minRetryDelay {
		t.Errorf("first delay %v", d)
	}
	if d := retryDelay(20*time.Second, errors.New("x")); d != maxRetryDelay {
		t.Errorf("capped delay %v", d)
	}
	if d := retryDelay(0, &Error{Code: 429, RetryAfter: 7}); d != 7*time.Second {
		t.Errorf("retry_after delay %v", d)
	}
}

func TestOptionalFieldsSerialization(t *testing.T) {
	cases := []struct {
		name   string
		params any
		want   string
	}{
		{"filter omitted", GetUpdatesParams{}, `{}`},
		{"filter reset", GetUpdatesParams{AllowedUpdates: []string{}}, `{"allowed_updates":[]}`},
		{"webhook filter reset", SetWebhookParams{URL: "u", AllowedUpdates: []string{}}, `{"url":"u","allowed_updates":[]}`},
		{"caption kept", CopyMessageParams{ChatID: 1, FromChatID: 2, MessageID: 3}, `{"chat_id":1,"from_chat_id":2,"message_id":3}`},
		{"caption removed", CopyMessageParams{ChatID: 1, FromChatID: 2, MessageID: 3, Caption: new(string)}, `{"chat_id":1,"from_chat_id":2,"message_id":3,"caption":""}`},
	}
	for _, c := range cases {
		got, err := json.Marshal(c.params)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}
