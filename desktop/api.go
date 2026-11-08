package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client talks to a Halo server's HTTP API: a session cookie, plus the CSRF
// token Halo also sets as a cookie, echoed in X-CSRF-Token on writes. Every
// response is {"data": …} or {"error": {"code", "message"}}, in snake_case.
type Client struct {
	mu      sync.Mutex
	base    string // "https://nas:7575", no trailing slash
	csrf    string
	pin     string // SHA-256 of a self-signed leaf certificate the user trusted
	http    *http.Client
	preview bool
}

// APIError is a failed request, with Halo's error code when it sent one.
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("HTTP %d", e.Status)
}

// UntrustedCertError is a server whose certificate the system does not trust
// and the user has not pinned: the UI shows the fingerprint and offers to
// trust it.
type UntrustedCertError struct {
	Fingerprint string
	Subject     string
}

func (e *UntrustedCertError) Error() string {
	return "服务器证书不受信任：" + e.Subject
}

func NewClient() *Client {
	c := &Client{}
	c.reset()
	return c
}

func (c *Client) reset() {
	jar, _ := cookiejar.New(nil)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		// Verified below: by the system's roots, else by the pinned
		// fingerprint. Halo on a NAS often serves its own self-signed
		// certificate, which the user trusts once, by its fingerprint.
		InsecureSkipVerify: true,
		VerifyConnection:   c.verify,
	}
	c.http = &http.Client{Jar: jar, Transport: transport, Timeout: 15 * time.Second}
	c.csrf = ""
}

func (c *Client) verify(cs tls.ConnectionState) error {
	if len(cs.PeerCertificates) == 0 {
		return errors.New("服务器没有提供证书")
	}
	leaf := cs.PeerCertificates[0]
	inter := x509.NewCertPool()
	for _, cert := range cs.PeerCertificates[1:] {
		inter.AddCert(cert)
	}
	if _, err := leaf.Verify(x509.VerifyOptions{DNSName: cs.ServerName, Intermediates: inter}); err == nil {
		return nil
	}
	sum := sha256.Sum256(leaf.Raw)
	fp := hex.EncodeToString(sum[:])
	c.mu.Lock()
	pin := c.pin
	c.mu.Unlock()
	if pin != "" && strings.EqualFold(pin, fp) {
		return nil
	}
	return &UntrustedCertError{Fingerprint: fp, Subject: leaf.Subject.CommonName}
}

// Configure points the client at a server, with the certificate pin the user
// trusted for it, and forgets the old session.
func (c *Client) Configure(server, pin string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.base = NormalizeServer(server)
	c.pin = pin
	c.preview = false
	c.reset()
}

func (c *Client) SetPreview(on bool) {
	c.mu.Lock()
	c.preview = on
	c.mu.Unlock()
}

func (c *Client) Preview() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.preview
}

func (c *Client) Base() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.base
}

// NormalizeServer turns what the user typed into a base URL: "nas:7575"
// becomes "http://nas:7575".
func NormalizeServer(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "/")
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "://") {
		s = "http://" + s
	}
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		u.Path = strings.TrimSuffix(u.Path, "/api")
		u.RawQuery, u.Fragment = "", ""
		return strings.TrimRight(u.String(), "/")
	}
	return s
}

// do sends a request to /api/<path> and decodes the response's data into out.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	return c.send(ctx, method, path, body, out, true)
}

func (c *Client) send(ctx context.Context, method, path string, body, out any, retry bool) error {
	c.mu.Lock()
	base, csrf, hc := c.base, c.csrf, c.http
	c.mu.Unlock()
	if base == "" {
		return errors.New("还没有设置服务器地址")
	}
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, base+"/api/"+strings.TrimPrefix(path, "/"), reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "HaloDesktop/"+appVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet && csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	resp, err := hc.Do(req)
	if err != nil {
		var cert *UntrustedCertError
		if errors.As(err, &cert) {
			return cert
		}
		return friendlyNetError(err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return err
	}
	var env struct {
		Data  json.RawMessage `json:"data"`
		Error json.RawMessage `json:"error"`
		Code  string          `json:"code"`
		Msg   string          `json:"message"`
	}
	_ = json.Unmarshal(raw, &env)
	if resp.StatusCode >= 400 {
		e := &APIError{Status: resp.StatusCode, Code: env.Code, Message: env.Msg}
		var detail struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(env.Error, &detail) == nil {
			if detail.Code != "" {
				e.Code = detail.Code
			}
			if detail.Message != "" {
				e.Message = detail.Message
			}
		} else {
			var s string
			if json.Unmarshal(env.Error, &s) == nil && s != "" {
				e.Message = s
			}
		}
		if e.Message == "" && resp.StatusCode == http.StatusUnauthorized {
			e.Message = "登录已过期，请重新登录"
		}
		// A stale CSRF token: fetch the session's and try once more.
		if retry && resp.StatusCode == http.StatusForbidden && e.Code == "invalid_csrf" {
			if _, err := c.Session(ctx); err == nil {
				return c.send(ctx, method, path, body, out, false)
			}
		}
		return e
	}
	if out == nil {
		return nil
	}
	data := env.Data
	if len(data) == 0 {
		data = raw
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("无法解析服务器响应（%s）：%w", path, err)
	}
	return nil
}

func friendlyNetError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "connection refused"):
		return errors.New("连接被拒绝：请检查地址和端口，Halo 默认端口是 7575")
	case strings.Contains(msg, "no such host"):
		return errors.New("找不到这个主机名")
	case strings.Contains(msg, "Client.Timeout") || strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "i/o timeout"):
		return errors.New("连接超时")
	case strings.Contains(msg, "server gave HTTP response to HTTPS client"):
		return errors.New("服务器没有启用 HTTPS，请把地址改成 http://")
	}
	return err
}

// ---- Session ----

type Session struct {
	Authenticated bool   `json:"authenticated"`
	CSRFToken     string `json:"csrf_token"`
	Username      string `json:"username"`
	Role          string `json:"role"`
	User          struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

func (s *Session) Name() string {
	if s.Username != "" {
		return s.Username
	}
	return s.User.Username
}

func (c *Client) remember(s *Session) {
	if s.CSRFToken != "" {
		c.mu.Lock()
		c.csrf = s.CSRFToken
		c.mu.Unlock()
	}
}

func (c *Client) Login(ctx context.Context, username, password string) (*Session, error) {
	var s Session
	if err := c.send(ctx, http.MethodPost, "auth/login", map[string]string{"username": username, "password": password}, &s, false); err != nil {
		return nil, err
	}
	c.remember(&s)
	return &s, nil
}

func (c *Client) Session(ctx context.Context) (*Session, error) {
	var s Session
	if err := c.send(ctx, http.MethodGet, "auth/session", nil, &s, false); err != nil {
		return nil, err
	}
	c.remember(&s)
	return &s, nil
}

func (c *Client) Logout(ctx context.Context) {
	_ = c.send(ctx, http.MethodPost, "auth/logout", map[string]any{}, nil, false)
	c.mu.Lock()
	c.reset()
	c.mu.Unlock()
}

// ---- Data ----

type SystemInfo struct {
	Version string `json:"version"`
}

func (c *Client) SystemInfo(ctx context.Context) (SystemInfo, error) {
	var v SystemInfo
	return v, c.do(ctx, http.MethodGet, "system/info", nil, &v)
}

func (c *Client) Host(ctx context.Context) (DashboardHost, error) {
	var v DashboardHost
	return v, c.do(ctx, http.MethodGet, "dashboard/host", nil, &v)
}

func (c *Client) Devices(ctx context.Context) ([]Device, error) {
	var v struct {
		Devices []Device `json:"devices"`
	}
	return v.Devices, c.do(ctx, http.MethodGet, "devices", nil, &v)
}

func (c *Client) Tasks(ctx context.Context) ([]Task, error) {
	var raw json.RawMessage
	if err := c.do(ctx, http.MethodGet, "automatic-tasks", nil, &raw); err != nil {
		return nil, err
	}
	var tasks []Task
	if json.Unmarshal(raw, &tasks) == nil {
		return tasks, nil
	}
	var wrapped struct {
		Tasks []Task `json:"tasks"`
	}
	err := json.Unmarshal(raw, &wrapped)
	return wrapped.Tasks, err
}

func (c *Client) Contacts(ctx context.Context) ([]Contact, error) {
	var v []Contact
	return v, c.do(ctx, http.MethodGet, "sms/contacts?limit=300", nil, &v)
}

func (c *Client) Thread(ctx context.Context, k Contact) ([]Message, error) {
	q := url.Values{"peer": {k.Peer}, "limit": {"200"}}
	if k.DeviceID != "" {
		q.Set("device_id", k.DeviceID)
	}
	if k.ModemIMEI != "" {
		q.Set("modem_imei", k.ModemIMEI)
	}
	if k.ICCID != "" {
		q.Set("iccid", k.ICCID)
	}
	if k.IMSI != "" {
		q.Set("imsi", k.IMSI)
	}
	var v []Message
	return v, c.do(ctx, http.MethodGet, "sms/thread?"+q.Encode(), nil, &v)
}

func (c *Client) SendSMS(ctx context.Context, deviceID, phone, text string) error {
	body := map[string]string{"phone": phone, "message": text}
	if deviceID != "" {
		body["device_id"] = deviceID
	}
	return c.do(ctx, http.MethodPost, "sms/send", body, nil)
}

func (c *Client) CallHistory(ctx context.Context) ([]CallRecord, error) {
	var v struct {
		Records []CallRecord `json:"records"`
	}
	return v.Records, c.do(ctx, http.MethodGet, "calls/history?limit=100", nil, &v)
}

func (c *Client) LiveCalls(ctx context.Context, deviceID string) ([]LiveCall, error) {
	var v struct {
		Calls []LiveCall `json:"calls"`
	}
	err := c.do(ctx, http.MethodGet, "devices/"+url.PathEscape(deviceID)+"/calls", nil, &v)
	for i := range v.Calls {
		v.Calls[i].DeviceID = deviceID
	}
	return v.Calls, err
}

func (c *Client) Dial(ctx context.Context, deviceID, number string) error {
	return c.do(ctx, http.MethodPost, "devices/"+url.PathEscape(deviceID)+"/calls/dial", map[string]string{"number": number}, nil)
}

func (c *Client) Answer(ctx context.Context, deviceID, callID string) error {
	return c.do(ctx, http.MethodPost, "devices/"+url.PathEscape(deviceID)+"/calls/answer", map[string]string{"call_id": callID}, nil)
}

func (c *Client) Hangup(ctx context.Context, deviceID, callID string) error {
	return c.do(ctx, http.MethodPost, "devices/"+url.PathEscape(deviceID)+"/calls/hangup", map[string]string{"call_id": callID}, nil)
}

func (c *Client) SetFlightMode(ctx context.Context, deviceID string, on bool) error {
	return c.do(ctx, http.MethodPatch, "devices/"+url.PathEscape(deviceID)+"/flight-mode", map[string]bool{"enabled": on}, nil)
}

func (c *Client) SetVoWiFi(ctx context.Context, deviceID string, on bool) error {
	return c.do(ctx, http.MethodPatch, "devices/"+url.PathEscape(deviceID)+"/vowifi", map[string]bool{"enabled": on}, nil)
}

func (c *Client) ReconnectVoWiFi(ctx context.Context, deviceID string) error {
	return c.do(ctx, http.MethodPost, "devices/"+url.PathEscape(deviceID)+"/vowifi/actions/reconnect", nil, nil)
}

func (c *Client) Reboot(ctx context.Context, deviceID string) error {
	return c.do(ctx, http.MethodPost, "devices/"+url.PathEscape(deviceID)+"/actions/reboot", nil, nil)
}

func (c *Client) Tunnels(ctx context.Context) (TunnelSnapshot, error) {
	var v TunnelSnapshot
	return v, c.do(ctx, http.MethodGet, "wireguard-tunnels", nil, &v)
}

func (c *Client) SetTunnel(ctx context.Context, id string, up bool) error {
	action := "down"
	if up {
		action = "up"
	}
	return c.do(ctx, http.MethodPost, "wireguard-tunnels/"+url.PathEscape(id)+"/"+action, nil, nil)
}

func (c *Client) Logs(ctx context.Context, level string) ([]LogEntry, error) {
	q := url.Values{"lines": {"500"}}
	if level != "" {
		q.Set("level", level)
	}
	var v struct {
		Logs []LogEntry `json:"logs"`
	}
	return v.Logs, c.do(ctx, http.MethodGet, "logs/history?"+q.Encode(), nil, &v)
}
