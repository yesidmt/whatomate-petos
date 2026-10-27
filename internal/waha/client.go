package waha

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const ExcludedPhone = "573104558226"

func isExcludedPhone(phone string) bool {
	return strings.HasSuffix(phone, "3104558226")
}

var phonePattern = regexp.MustCompile(`^[1-9][0-9]{7,14}$`)

// Client is an optional, single-session temporary transport. Its API key is
// deliberately kept in the process environment, never in an account record.
type Client struct {
	BaseURL       string
	APIKey        string
	Session       string
	ExpectedPhone string
	HTTP          *http.Client
}

func NormalizePhone(value string) string {
	var b strings.Builder
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func NewFromEnv() (*Client, error) {
	base := strings.TrimSpace(os.Getenv("PETOS_WAHA_URL"))
	if base == "" {
		return nil, nil
	}
	key := strings.TrimSpace(os.Getenv("PETOS_WAHA_API_KEY"))
	session := strings.TrimSpace(os.Getenv("PETOS_WAHA_SESSION"))
	phone := NormalizePhone(os.Getenv("PETOS_WAHA_PHONE"))
	if key == "" || session == "" || !phonePattern.MatchString(phone) {
		return nil, errors.New("PETOS_WAHA_API_KEY, PETOS_WAHA_SESSION and a valid PETOS_WAHA_PHONE are required")
	}
	if isExcludedPhone(phone) {
		return nil, errors.New("the phone ending in 8226 is excluded from WAHA")
	}
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("PETOS_WAHA_URL must be an HTTP(S) origin")
	}
	if u.Scheme == "http" && u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" && u.Hostname() != "waha" {
		return nil, errors.New("remote WAHA requires HTTPS")
	}
	return &Client{BaseURL: strings.TrimRight(base, "/"), APIKey: key, Session: session, ExpectedPhone: phone, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *Client) request(ctx context.Context, method, path string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("WAHA returned HTTP %d", resp.StatusCode)
	}
	return data, nil
}

// VerifySession fails closed unless WAHA confirms the exact, approved sender.
// It is called before every send so reconnecting a different phone cannot
// silently redirect traffic from the same Whatomate account.
func (c *Client) VerifySession(ctx context.Context) error {
	data, err := c.request(ctx, http.MethodGet, "/api/sessions/"+url.PathEscape(c.Session), nil)
	if err != nil {
		return err
	}
	var session struct {
		Status string `json:"status"`
		Me     *struct {
			ID string `json:"id"`
		} `json:"me"`
	}
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("invalid WAHA session response: %w", err)
	}
	if session.Status != "WORKING" || session.Me == nil {
		return errors.New("WAHA session is not connected")
	}
	actual := NormalizePhone(strings.SplitN(session.Me.ID, "@", 2)[0])
	if isExcludedPhone(actual) || actual != c.ExpectedPhone {
		return errors.New("WAHA is connected to a different or excluded phone")
	}
	return nil
}

func (c *Client) SendText(ctx context.Context, phone, message, replyTo string) (string, error) {
	phone = NormalizePhone(phone)
	if !phonePattern.MatchString(phone) || isExcludedPhone(phone) {
		return "", errors.New("invalid or excluded recipient phone")
	}
	if strings.TrimSpace(message) == "" {
		return "", errors.New("message text is required")
	}
	if err := c.VerifySession(ctx); err != nil {
		return "", err
	}
	payload := map[string]any{"session": c.Session, "chatId": phone + "@c.us", "text": message}
	if replyTo != "" {
		payload["reply_to"] = replyTo
	}
	data, err := c.request(ctx, http.MethodPost, "/api/sendText", payload)
	if err != nil {
		return "", err
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.ID == "" {
		return "", errors.New("WAHA did not return a message ID")
	}
	return result.ID, nil
}
