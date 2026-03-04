package notifier

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
}

type Client struct {
	http   *http.Client
	config Config
}

type Message struct {
	Title    string
	Priority int
	Tags     []string
	Text     string
}

func NewClient(config Config) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
		config: config,
	}
}

func (c *Client) SendMessage(ctx context.Context, topic string, message Message) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/"+topic, strings.NewReader(message.Text))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Title", message.Title)
	req.Header.Set("Priority", fmt.Sprintf("%d", message.Priority))
	req.Header.Set("Tags", strings.Join(message.Tags, ","))

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
