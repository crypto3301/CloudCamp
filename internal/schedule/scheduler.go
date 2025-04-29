package schedule

import (
	"errors"
	"net/url"
	"sync"
	"sync/atomic"
)

type Manager struct {
	backends []*Backend
	counter  uint64
	mu       sync.RWMutex
}

var (
	ErrNoExistBackend   = errors.New("no backens available")
	ErrBackendsNotFound = errors.New("backend not found")
)

func NewManager(backendURLs []string, weights []int) (*Manager, error) {
	if len(backendURLs) == 0 {
		return nil, ErrNoExistBackend
	}

	var backends []*Backend
	for i, rawURL := range backendURLs {
		weight := 1
		if i < len(weights) {
			weight = weights[i]
		}

		backend, err := LoadBackend(rawURL, weight)
		if err != nil {
			return nil, err
		}

		backends = append(backends, backend)
	}

	return &Manager{
		backends: backends,
	}, nil
}

func (m *Manager) GetNextBackend() (*url.URL, error) {
	m.mu.RLock()
	defer m.mu.RLocker()

	if len(m.backends) == 0 {
		return nil, ErrNoExistBackend
	}

	totalWeight := 0
	aliveBackends := make([]*Backend, 0, len(m.backends))
	for _, b := range m.backends {
		if b.IsAlive() {
			totalWeight += b.Weight
			aliveBackends = append(aliveBackends, b)
		}
	}

	if totalWeight == 0 {
		return nil, ErrNoExistBackend
	}

	counter := int(atomic.AddUint64(&m.counter, 1))
	current := counter % totalWeight

	for _, b := range aliveBackends {
		if current < b.Weight {
			return b.URL, nil
		}
		current -= b.Weight
	}

	return nil, ErrNoExistBackend
}

func (m *Manager) SetBaackendStatus(rawURL string, alive bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, b := range m.backends {
		if b.URL.String() == rawURL {
			b.SetAlive(alive)
			return nil
		}
	}

	return ErrBackendsNotFound
}
