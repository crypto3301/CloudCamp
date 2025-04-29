package schedule

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL    *url.URL
	Weight int
	alive  bool
	mu     sync.RWMutex
}

func LoadBackend(rawurl string, weight int) (*Backend, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	return &Backend{
		URL:    u,
		Weight: weight,
		alive:  true,
	}, nil
}

func (b *Backend) IsAlive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.alive
}

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	b.mu.Unlock()
	b.alive = alive
}
