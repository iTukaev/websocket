package main

import (
	"betapi_server/config/trconfig"
	"betapi_server/provider"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	ProviderPassKey = "?????????????????"
)

type Sender interface {
	sendMessage(b []byte)
}

// Manager manages requests and checks body,
// stores the latest modified version of provider's response.
type Manager struct {
	sender    Sender
	transport *http.Transport
	timestamp time.Time
	ticker    *time.Ticker
	url       string
	sync.Mutex
	Body *provider.Matches
}

func NewManager(sender Sender, tick time.Duration, url string, dataType string) *Manager {
	return &Manager{
		sender:    sender,
		timestamp: time.Now(),
		ticker:    time.NewTicker(tick),
		transport: trconfig.Transport(),
		url:       url,
		Body:      provider.NewMatches(dataType),
	}
}

// Load structure for current response data
type Load struct {
	timestamp time.Time
	Body      *provider.Matches
}

func NewLoad(dataType string) *Load {
	return &Load{
		timestamp: time.Now(),
		Body:      provider.NewMatches(dataType),
	}
}

// checkNewMatches start dataFromProvider with ticker time period,
// stopped work by context.Done is checked
func (m *Manager) checkNewMatches(ctx context.Context, logger *zap.SugaredLogger) {
	go m.dataFromProvider(logger)
Loop:
	for {
		select {
		case <-m.ticker.C:
			go m.dataFromProvider(logger)
		case <-ctx.Done():
			time.Sleep(3 * time.Second)
			break Loop
		}
	}
}

// dataFromProvider make request to provider
// if changes are checked, spend new data to clients
func (m *Manager) dataFromProvider(logger *zap.SugaredLogger) {
	client := trconfig.Client(m.transport)
	inputLoad := NewLoad(m.Body.Parameter)

	req, err := http.NewRequest(http.MethodGet, m.url, nil)
	if err != nil {
		logger.Fatalf("new request creating error: %v", err)
	}
	req.Header.Set("Package", ProviderPassKey)

	resp, err := client.Do(req)
	if err != nil {
		logger.Fatalf("response error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalf("reading body error: %v", err)
	}

	if err := json.Unmarshal(respBody, inputLoad.Body); err != nil {
		logger.Errorf("unmarshalling error: %v", err)
		return
	}

	if !m.timestamp.Before(inputLoad.timestamp) {
		logger.Warn("New data timestamp is before sample timestamp ")
		return
	}

	inputLoad.Body.Sort()
	m.Lock()
	b, err := inputLoad.Body.Bytes()
	if err != nil {
		logger.Error(err)
		return
	}

	if !m.Body.Equal(inputLoad.Body) {
		m.Body = inputLoad.Body
		m.sender.sendMessage(b)
	}

	m.timestamp = inputLoad.timestamp
	m.Unlock()
}
