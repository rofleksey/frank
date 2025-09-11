package yandex

import (
	"frank/pkg/config"
	"net/http"
	"time"

	"github.com/samber/do"
)

type Client struct {
	cfg    *config.Config
	client *http.Client
}

func NewClient(di *do.Injector) (*Client, error) {
	cfg := do.MustInvoke[*config.Config](di)

	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}, nil
}
