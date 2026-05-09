package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	host          string
	username      string
	password      string
	debugResponse bool
	httpClient    *http.Client
	mu            sync.Mutex
}

func NewClient(host string, skipTLS bool, username, password string, debugResponse bool) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("parse host: %w", err)
	}

	jar.SetCookies(baseURL, []*http.Cookie{
		{Name: "language", Value: "English"},
	})

	return &Client{
		host:          strings.TrimRight(host, "/"),
		username:      username,
		password:      password,
		debugResponse: debugResponse,
		httpClient: &http.Client{
			Jar: jar,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS}, //nolint:gosec
			},
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: 15 * time.Second,
		},
	}, nil
}

func (c *Client) Login(ctx context.Context) error {
	form := url.Values{
		"user_name":     []string{encodeCredential(c.username)},
		"user_password": []string{encodeCredential(c.password)},
		"lan":           []string{"en"},
		"op_Type":       []string{"1"},
		"rand_code":     []string{"0"},
		"tokenID":       []string{"$[$ID_TOKEN_ID]"},
		"validateValue": []string{"0"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.host+"/cgi-bin/login.cgi", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && (resp.StatusCode < 300 || resp.StatusCode >= 400) {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) KeepAlive(ctx context.Context) error {
	_, err := c.do(ctx, http.MethodGet, "/cgi-bin/main_page_polling.cgi", nil)
	return err
}

func (c *Client) FetchDeviceData(ctx context.Context, equipID int) (map[int]Sample, error) {
	query := url.Values{
		"_equipId": []string{fmt.Sprintf("%d", equipID)},
		"_op_type": []string{"1"},
		"sand":     []string{fmt.Sprintf("%.16f", rand.Float64())},
	}

	body, err := c.do(ctx, http.MethodGet, "/cgi-bin/p05_equip_sample.cgi", query)
	if err != nil {
		return nil, err
	}

	samples, err := ParseSamples(string(body))
	if err != nil {
		if c.debugResponse {
			return nil, fmt.Errorf("%w; full response: %s", err, strconv.Quote(strings.TrimSpace(string(body))))
		}
		return nil, err
	}

	return samples, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.doLocked(ctx, method, path, query, true)
}

func (c *Client) doLocked(ctx context.Context, method, path string, query url.Values, allowRelogin bool) ([]byte, error) {
	endpoint := c.host + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusFound {
		if !allowRelogin {
			return nil, fmt.Errorf("authentication failed with status %d", resp.StatusCode)
		}
		if err := c.Login(ctx); err != nil {
			return nil, err
		}
		return c.doLocked(ctx, method, path, query, false)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return body, nil
}

func encodeCredential(raw string) string {
	data := []byte(raw)
	if len(data) < 9 {
		padded := make([]byte, 9)
		copy(padded, data)
		data = padded
	}
	return base64.RawStdEncoding.EncodeToString(data)
}
