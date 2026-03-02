package untis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type Client struct {
	http   *http.Client
	config Config
}

func NewClient(config Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookie jar %w", err)
	}
	return &Client{
		http: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
		config: config,
	}, nil
}

func extractSchoolID(cookies []*http.Cookie) (string, error) {
	for _, c := range cookies {
		if c.Name == "schoolname" {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("login failed: missing schoolname cookie")
}
func login(username, password string, config Config) (Session, error) {
	resp, err := client.PostForm(config.BaseURL+"/WebUntis/j_spring_security_check", url.Values{
		"school":     {config.SchoolName},
		"j_username": {username},
		"j_password": {password},
		"token":      {""},
	})
	if err != nil {
		return Session{}, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Session{}, fmt.Errorf("reading login response: %w", err)
	}
	var raw LoginResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return Session{}, fmt.Errorf("parsing login response: %w", err)
	}
	if raw.Result == nil {
		msg := "login failed"
		if raw.Error != nil {
			msg = raw.Error.Message
		}
		return Session{}, fmt.Errorf(msg)
	}
	schoolID, err := extractSchoolID(resp.Cookies())
	if err != nil {
		return Session{}, err
	}
	return Session{
		SessionID: raw.Result.SessionID,
		Token:     raw.Result.Token,
		SchoolID:  schoolID,
	}, nil
}

func getStaticInfo(session Session) UntisInfo {
	cli
}
