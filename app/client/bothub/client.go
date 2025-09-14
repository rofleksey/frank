package bothub

import (
	"fmt"
	"frank/pkg/config"
	"net"
	"net/http"
	"time"

	"github.com/samber/do"
)

var timeout = 10 * time.Minute

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
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   timeout,
					KeepAlive: timeout,
				}).DialContext,
				TLSHandshakeTimeout:   timeout,
				ResponseHeaderTimeout: timeout,
				ExpectContinueTimeout: timeout,
			},
		},
		token:   cfg.Bothub.Token,
		baseURL: "https://bothub.chat/api/v2/openai/v1/chat/completions",
	}, nil
}
