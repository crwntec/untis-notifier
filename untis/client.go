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
			Jar:           jar,
			Timeout:       15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
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
func (c *Client) Login(username, password string) (Session, error) {
	resp, err := c.http.PostForm(c.config.BaseURL+"/WebUntis/j_spring_security_check", url.Values{
		"school":     {c.config.SchoolName},
		"j_username": {username},
		"j_password": {password},
		"token":      {""},
	})
	if err != nil {
		return Session{}, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if resp.StatusCode != http.StatusFound || location != "/WebUntis/index.do" {
		return Session{}, fmt.Errorf("login failed: unexpected redirect to %q (status %d)", location, resp.StatusCode)
	}
	schoolID, err := extractSchoolID(resp.Cookies())
	if err != nil {
		return Session{}, err
	}
	var jsessionID string
	for _, c := range resp.Cookies() {
		if c.Name == "JSESSIONID" {
			jsessionID = c.Value
			break
		}
	}
	if jsessionID == "" {
		return Session{}, fmt.Errorf("login failed: missing JSESSIONID cookie")
	}
	req, err := http.NewRequest("GET", c.config.BaseURL+"/WebUntis/api/token/new", nil)

	if err != nil {
		return Session{}, fmt.Errorf("login failed: %w", err)
	}
	req.Header.Add("Cookie", fmt.Sprintf("JSESSIONID=%s;", jsessionID))

	res, err := c.http.Do(req)
	if err != nil {
		return Session{}, fmt.Errorf("login failed: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Session{}, fmt.Errorf("login failed: %w", err)
	}
	var token = string(body)

	return Session{
		SessionID: jsessionID,
		SchoolID:  schoolID,
		Token:     token,
	}, nil
}

func (c *Client) GetStaticInfo(session Session) (UntisInfo, error) {
	req, err := http.NewRequest("GET", c.config.BaseURL+"/WebUntis/api/rest/view/v1/app/data", nil)
	if err != nil {
		return UntisInfo{}, fmt.Errorf("creating request failed %w", err)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+session.Token)
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
		SchoolID:          session.SchoolID,
		CurrentSchoolYear: appData.CurrentSchoolYear,
		Holidays:          appData.Holidays,
	}, nil
}

func (c *Client) GetTimetable(info UntisInfo, session Session, start, end string) (Timetable, error) {
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
	req.Header.Set("Authorization", "Bearer "+session.Token)
	resp, err := c.http.Do(req)
	if err != nil {
		return Timetable{}, fmt.Errorf("fetching timetable failed %w", err)
	}
	var timetableResponse TimetableResponse
	if err := json.NewDecoder(resp.Body).Decode(&timetableResponse); err != nil {
		return Timetable{}, fmt.Errorf("parsing timetable failed %w", err)
	}
	return TimetableFromResponse(timetableResponse)
}
