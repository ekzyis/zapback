package lnurl

import (
	"fmt"
	"net/http"
	"strings"
)

func VerifyLNURLp(address string) error {
	parts := strings.Split(address, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid lightning address format: %s", address)
	}
	name := parts[0]
	domain := parts[1]

	url := fmt.Sprintf("https://%s/.well-known/lnurlp/%s", domain, name)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lnurlp: expected 200 as status code, got %s", resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return fmt.Errorf("lnurlp: expected application/json as content type, got %s", contentType)
	}
	return nil
}
