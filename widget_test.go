package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSendWidget(t *testing.T) {
	fs := newFakeServer(t)
	fs.handle("sendWidget", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if got["chat_id"] != float64(42) || got["html"] != "<p>hi</p>" || got["allow_queries"] != true || got["chat_data"] != "{}" {
			t.Errorf("unexpected params %v", got)
		}
		for _, k := range []string{"width", "height", "public_data", "is_shared"} {
			if _, ok := got[k]; ok {
				t.Errorf("%s should be omitted when zero", k)
			}
		}
		reply(w, `{"ok":true,"result":{"message_id":7,"date":1,"chat":{"id":42,"type":"private"},"widget":{"html":"<p>hi</p>","chat_data":"{}"}}}`)
	})

	msg, err := fs.api().SendWidget(context.Background(), SendWidgetParams{
		ChatID: 42, HTML: "<p>hi</p>", ChatData: "{}", AllowQueries: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.MessageID != 7 || msg.Widget == nil || msg.Widget.ChatData != "{}" {
		t.Errorf("unexpected message %+v", msg)
	}
}

func TestEditMessageWidgetDataKeepsNilFields(t *testing.T) {
	fs := newFakeServer(t)
	fs.handle("editMessageWidgetData", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if _, ok := got["public_data"]; ok {
			t.Errorf("nil public_data must be omitted (left unchanged)")
		}
		if v, ok := got["chat_data"]; !ok || v != "" {
			t.Errorf("empty chat_data must be sent to clear it, got %v", got)
		}
		reply(w, `{"ok":true,"result":true}`)
	})
	empty := ""
	if err := fs.api().EditMessageWidgetData(context.Background(), EditMessageWidgetDataParams{ChatID: 1, MessageID: 2, ChatData: &empty}); err != nil {
		t.Fatal(err)
	}
}

func TestWidgetQueryDecodingAndOrderKey(t *testing.T) {
	var upd Update
	raw := `{"update_id":1,"widget_query":{"id":"q","from":{"id":5,"is_bot":false,"first_name":"A"},
		"chat":{"id":5,"type":"private"},"message_id":9,"chat_instance":"c","type":{"type":"data"},"data":"p"}}`
	if err := json.Unmarshal([]byte(raw), &upd); err != nil {
		t.Fatal(err)
	}
	q := upd.WidgetQuery
	if q == nil || q.ID != "q" || q.MessageID != 9 || q.Type.Type != WidgetQueryData || q.Data != "p" {
		t.Fatalf("unexpected query %+v", q)
	}
	if id, ok := UpdateChatID(&upd); !ok || id != 5 {
		t.Errorf("UpdateChatID = %d, %v", id, ok)
	}
	shared := Update{SharedWidgetQuery: &SharedWidgetQuery{ID: "s", From: &User{ID: 6}, WidgetID: "w"}}
	if id, ok := UpdateChatID(&shared); !ok || id != 6 {
		t.Errorf("UpdateChatID(shared) = %d, %v", id, ok)
	}
}
