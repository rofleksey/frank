package bothub

import (
	"fmt"
	"frank/pkg/config"
	"net"
	"net/http"
	"time"

	"github.com/samber/do"
)

// Client represents the Bothub chat API client
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// NewClient creates a new Bothub client instance
func NewClient(di *do.Injector) (*Client, error) {
	cfg := do.MustInvoke[*config.Config](di)

	if cfg.Bothub.Token == "" {
		return nil, fmt.Errorf("bothub token is required")
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Minute,
					KeepAlive: 5 * time.Minute,
				}).DialContext,
				TLSHandshakeTimeout:   5 * time.Minute,
				ResponseHeaderTimeout: 5 * time.Minute,
				ExpectContinueTimeout: 5 * time.Minute,
			},
		},
		token:   cfg.Bothub.Token,
		baseURL: "https://bothub.chat/api/v2/openai/v1/chat/completions",
	}, nil
}
