package httpclient

import (
	"html"
	"strings"
)

func formatURL(url string) string {
	url = html.UnescapeString(url)
	if !strings.HasPrefix(url, "http") {
		return DefaultProtocol + "://" + url
	}
	return url
}
