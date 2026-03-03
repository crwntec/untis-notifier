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
	http     *http.Client
	config   Config
	Session  Session
	username string
	password string
}

func NewClient(config Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookie jar %w", err)
	}
	return &Client{
		http: &http.Client{
			Jar:           jar,
			Timeout:       15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		},
		config: config,
	}, nil
}

func (c *Client) newAuthedRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new authed request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Session.Token)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	return req, nil
}

func extractSchoolID(cookies []*http.Cookie) (string, error) {
	for _, c := range cookies {
		if c.Name == "schoolname" {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("login failed: missing schoolname cookie")
}

func (c *Client) refreshToken() error {
	req, err := http.NewRequest("GET", c.config.BaseURL+"/WebUntis/api/token/new", nil)
	if err != nil {
		return fmt.Errorf("refresh token failed: %w", err)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("refresh token failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return c.Login(c.username, c.password)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("refresh token failed: %w", err)
	}
	c.Session.Token = string(body)
	return nil
}

func (c *Client) Login(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("login failed: username or password empty")
	}
	resp, err := c.http.PostForm(c.config.BaseURL+"/WebUntis/j_spring_security_check", url.Values{
		"school":     {c.config.SchoolName},
		"j_username": {username},
		"j_password": {password},
		"token":      {""},
	})
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if resp.StatusCode != http.StatusFound || location != "/WebUntis/index.do" {
		if location == "/WebUntis/" {
			return fmt.Errorf("login failed: invalid credentials")
		}
		return fmt.Errorf("login failed: unexpected redirect to %q (status %d)", location, resp.StatusCode)
	}
	schoolID, err := extractSchoolID(resp.Cookies())
	if err != nil {
		return err
	}
	var jsessionID string
	for _, c := range resp.Cookies() {
		if c.Name == "JSESSIONID" {
			jsessionID = c.Value
			break
		}
	}
	if jsessionID == "" {
		return fmt.Errorf("login failed: missing JSESSIONID cookie")
	}
	u, err := url.Parse(c.config.BaseURL + "/WebUntis")
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	c.http.Jar.SetCookies(u, resp.Cookies())
	c.Session = Session{
		SessionID: jsessionID,
		SchoolID:  schoolID,
	}
	c.username = username
	c.password = password
	err = c.refreshToken()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetStaticInfo() (UntisInfo, error) {
	req, err := c.newAuthedRequest("GET", c.config.BaseURL+"/WebUntis/api/rest/view/v1/app/data")
	if err != nil {
		return UntisInfo{}, fmt.Errorf("creating request failed %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return UntisInfo{}, fmt.Errorf("getting info failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return UntisInfo{}, fmt.Errorf("app data request returned status %d", resp.StatusCode)
	}

	var appData AppData
	if err := json.NewDecoder(resp.Body).Decode(&appData); err != nil {
		return UntisInfo{}, fmt.Errorf("parsing app data failed: %w", err)
	}

	return UntisInfo{
		UserID:            appData.User.Person.ID,
		SchoolID:          c.Session.SchoolID,
		CurrentSchoolYear: appData.CurrentSchoolYear,
		Holidays:          appData.Holidays,
	}, nil
}

func (c *Client) GetTimetable(info UntisInfo, start, end string) (Timetable, error) {
	req, err := http.NewRequest("GET", c.config.BaseURL+"/WebUntis/api/rest/view/v1/timetable/entries", nil)
	if err != nil {
		return Timetable{}, fmt.Errorf("creating request failed %w", err)
	}
	q := req.URL.Query()
	q.Add("start", start)
	q.Add("end", end)
	q.Add("format", "8")
	q.Add("resourceType", "STUDENT")
	q.Add("resources", fmt.Sprintf("%d", info.UserID))
	q.Add("periodTypes", "")
	q.Add("timetableType", "MY_TIMETABLE")
	q.Add("layout", "START_TIME")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+c.Session.Token)
	resp, err := c.http.Do(req)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			if err := c.refreshToken(); err != nil {
				return Timetable{}, fmt.Errorf("fetching timetable failed %w", err)
			}
			return c.GetTimetable(info, start, end)
		}
		return Timetable{}, fmt.Errorf("fetching timetable failed %w", err)
	}
	var timetableResponse TimetableResponse
	if err := json.NewDecoder(resp.Body).Decode(&timetableResponse); err != nil {
		return Timetable{}, fmt.Errorf("parsing timetable failed %w", err)
	}
	return TimetableFromResponse(timetableResponse)
}
