package telegram

import (
	"fmt"
	"github.com/minya/goutils/web"
	"net/http"
)

func ReadUpdates(botId string, offset int, limit int) ([]Update, error) {
	query := ""
	if offset > 0 {
		query += fmt.Sprintf("offset=%v", offset)
	}
	if limit > 0 {
		if query != "" {
			query += "&"
		}
		query += fmt.Sprintf("limit=%v", offset)
	}
	url := fmt.Sprint("https://api.telegram.org/bot%v/getUpdates", botId)
	if query != "" {
		url += "?" + query
	}

	client := http.Client{
		Transport: web.DefaultTransport(1000),
	}

	_, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
