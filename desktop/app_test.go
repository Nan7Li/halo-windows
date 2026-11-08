package main

import (
	"encoding/json"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// testApp runs everything synchronously: the view, the network and the
// updates, so a test reads the outcome right after the input.
func testApp(t *testing.T) (*app, *memPrefs) {
	t.Helper()
	store := &memPrefs{p: defaultPrefs()}
	a := newApp(store)
	a.now = func() time.Time { return time.Date(2026, 11, 8, 15, 30, 0, 0, time.Local) }
	a.spawn = func(fn func()) { fn() }
	a.start()
	return a, store
}

// clickRole clicks the middle of the element with that role and name, for
// names that other elements show as text too.
func clickRole(t *testing.T, tt *ui.Tester, role ui.Role, name string) {
	t.Helper()
	info, ok := tt.AccessByRole(role, name)
	if !ok {
		t.Fatalf("no %v named %q", role, name)
	}
	b := info.Bounds
	tt.ClickAt(b.X+b.W/2, b.Y+b.H/2)
}

// snapshot writes the tester's last frame to $SNAPSHOT_DIR/name.png.
func snapshot(t *testing.T, tt *ui.Tester, name string) {
	dir := os.Getenv("SNAPSHOT_DIR")
	if dir == "" {
		return
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}

func TestPreviewWalksEveryPage(t *testing.T) {
	for _, dark := range []bool{false, true} {
		a, _ := testApp(t)
		tt := ui.NewTesterWith(a.view, 1120, 740, ui.TesterOptions{Dark: dark})
		mode := map[bool]string{false: "light", true: "dark"}[dark]
		if !tt.HasText("登录 Halo") {
			t.Fatalf("login screen texts %q", tt.Texts())
		}
		snapshot(t, tt, "login-"+mode)
		if err := tt.Click("不登录，先看看界面"); err != nil {
			t.Fatal(err)
		}
		want := map[string]string{
			pageToday:    "Intel Celeron J3160",
			pageMessages: "Alex Chen",
			pagePhone:    "最近通话",
			pageDevices:  "EC25 · 主卡",
			pageNetwork:  "home-core",
			pageLogs:     "IMS registered on Vodafone UK ePDG",
			pageSettings: "强调色",
		}
		for _, pg := range pages {
			if err := tt.Click(pg.title); err != nil {
				t.Fatalf("%s: %v", pg.title, err)
			}
			tt.Advance(time.Second)
			if a.page != pg.id {
				t.Fatalf("clicked %s, page is %s", pg.title, a.page)
			}
			if !tt.HasText(want[pg.id]) {
				t.Errorf("%s (%s) lacks %q: %q", pg.id, mode, want[pg.id], tt.Texts())
			}
			snapshot(t, tt, pg.id+"-"+mode)
		}
	}
}

func TestPreviewConversationSends(t *testing.T) {
	a, _ := testApp(t)
	tt := ui.NewTester(a.view, 1120, 740)
	if err := tt.Click("不登录，先看看界面"); err != nil {
		t.Fatal(err)
	}
	tt.Click("信息")
	if err := tt.Click("Alex Chen"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	if !tt.HasText("还在滑行，等下打给你") {
		t.Fatalf("thread not shown: %q", tt.Texts())
	}
	if err := tt.Click("短信内容"); err != nil {
		t.Fatal(err)
	}
	tt.Type("晚上见")
	tt.Key(0, ui.KeyEnter)
	tt.Advance(time.Second)
	if !tt.HasText("晚上见") || a.draft != "" {
		t.Errorf("message not sent: draft %q, texts %q", a.draft, tt.Texts())
	}
	snapshot(t, tt, "conversation-sent")
}

func TestKeypadDials(t *testing.T) {
	a, _ := testApp(t)
	tt := ui.NewTester(a.view, 1120, 740)
	tt.Click("不登录，先看看界面")
	tt.Click("电话")
	for _, d := range []string{"1", "0", "0", "8", "6"} {
		clickRole(t, tt, ui.RoleButton, d)
	}
	if a.dialed != "10086" {
		t.Fatalf("dialed %q", a.dialed)
	}
	clickRole(t, tt, ui.RoleButton, "删除")
	if a.dialed != "1008" {
		t.Fatalf("after delete %q", a.dialed)
	}
	snapshot(t, tt, "keypad")
}

func TestSettingsChangeAppearance(t *testing.T) {
	a, store := testApp(t)
	tt := ui.NewTester(a.view, 1120, 740)
	tt.Click("不登录，先看看界面")
	tt.Click("设置")
	if err := tt.Click("深色"); err != nil {
		t.Fatal(err)
	}
	if store.p.Appearance != "dark" {
		t.Errorf("appearance %q", store.p.Appearance)
	}
	if err := tt.Click("紫色"); err != nil {
		t.Fatal(err)
	}
	if store.p.Accent != "purple" {
		t.Errorf("accent %q", store.p.Accent)
	}
	snapshot(t, tt, "settings-dark-purple")
}

func TestIncomingCallBanner(t *testing.T) {
	a, _ := testApp(t)
	var notes []string
	a.notify = func(title, body, id string) { notes = append(notes, title+":"+body) }
	tt := ui.NewTester(a.view, 1120, 740)
	tt.Click("不登录，先看看界面")
	a.setLive([]LiveCall{{DeviceID: "ec25", ID: "c1", Number: "+12025550123", Direction: "incoming", State: "ringing", StartedAt: a.now()}})
	tt.Frame()
	if !tt.HasText("来电") || len(notes) != 1 {
		t.Fatalf("banner/notification missing: notes %q texts %q", notes, tt.Texts())
	}
	snapshot(t, tt, "incoming-call")
	a.setLive([]LiveCall{{DeviceID: "ec25", ID: "c1", Number: "+12025550123", Direction: "incoming", State: "ringing", StartedAt: a.now()}})
	if len(notes) != 1 {
		t.Errorf("notified twice for one call: %q", notes)
	}
}

func TestNotifiesOnlyNewMessages(t *testing.T) {
	a, _ := testApp(t)
	var notes []string
	a.notify = func(title, body, id string) { notes = append(notes, title) }
	a.phase = phaseSignedIn
	a.contacts = []Contact{{Peer: "A", LastSMSID: 1, UnreadCount: 1}}
	a.notifyMessages() // baseline
	if len(notes) != 0 {
		t.Fatalf("notified on first load: %q", notes)
	}
	a.contacts = []Contact{{Peer: "A", LastSMSID: 2, UnreadCount: 2, LastMessage: "hi"}, {Peer: "B", LastSMSID: 5, Direction: "outbound"}}
	a.notifyMessages()
	if len(notes) != 1 || notes[0] != "A" {
		t.Errorf("notes %q", notes)
	}
}

// fakeHalo is enough of Halo's API for the client: a session cookie, the
// double-submit CSRF token, and a few endpoints, in snake_case.
type fakeHalo struct {
	mu      sync.Mutex
	csrf    string
	sent    []map[string]string
	rotated bool
}

func (f *fakeHalo) handler(t *testing.T) http.Handler {
	mux := http.NewServeMux()
	write := func(w http.ResponseWriter, status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	authed := func(w http.ResponseWriter, r *http.Request) bool {
		if c, err := r.Cookie("vocat_session"); err != nil || c.Value != "s1" {
			write(w, 401, map[string]any{"error": map[string]string{"code": "unauthorized", "message": "authentication required"}})
			return false
		}
		if r.Method != http.MethodGet {
			f.mu.Lock()
			want := f.csrf
			f.mu.Unlock()
			cookie, _ := r.Cookie("vocat_csrf")
			if r.Header.Get("X-CSRF-Token") != want || cookie == nil || cookie.Value != want {
				write(w, 403, map[string]any{"error": map[string]string{"code": "invalid_csrf", "message": "CSRF validation failed"}})
				return false
			}
		}
		return true
	}
	setCSRF := func(w http.ResponseWriter, token string) {
		f.mu.Lock()
		f.csrf = token
		f.mu.Unlock()
		http.SetCookie(w, &http.Cookie{Name: "vocat_csrf", Value: token, Path: "/"})
	}
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["username"] != "admin" || body["password"] != "secret" {
			write(w, 401, map[string]any{"error": map[string]string{"code": "invalid_credentials", "message": "用户名或密码错误"}})
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "vocat_session", Value: "s1", Path: "/", HttpOnly: true})
		setCSRF(w, "t1")
		write(w, 200, map[string]any{"data": map[string]any{"csrf_token": "t1", "authenticated": true, "user": map[string]string{"username": "admin"}}})
	})
	mux.HandleFunc("/api/auth/session", func(w http.ResponseWriter, r *http.Request) {
		if !authed(w, r) {
			return
		}
		f.mu.Lock()
		token := f.csrf
		f.mu.Unlock()
		write(w, 200, map[string]any{"data": map[string]any{"csrf_token": token, "authenticated": true, "username": "admin"}})
	})
	mux.HandleFunc("/api/system/info", func(w http.ResponseWriter, r *http.Request) {
		if authed(w, r) {
			write(w, 200, map[string]any{"data": map[string]string{"version": "1.1.15"}})
		}
	})
	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		if authed(w, r) {
			write(w, 200, map[string]any{"data": map[string]any{"device_limit": 4, "devices": []map[string]any{{
				"id": "ec25", "name": "EC25", "running": true, "healthy": true, "vowifi_enabled": true,
				"vowifi_runtime": map[string]any{"ims_ready": true}, "modem": map[string]any{"operator": "Vodafone UK", "signal_dbm": -90, "network_mode": "LTE"},
			}}}})
		}
	})
	mux.HandleFunc("/api/dashboard/host", func(w http.ResponseWriter, r *http.Request) {
		if authed(w, r) {
			write(w, 200, map[string]any{"data": map[string]any{"host": map[string]string{"board_model": "QDC507"}, "perf": map[string]float64{"cpu_percent": 12}}})
		}
	})
	mux.HandleFunc("/api/sms/contacts", func(w http.ResponseWriter, r *http.Request) {
		if authed(w, r) {
			write(w, 200, map[string]any{"data": []map[string]any{{"device_id": "ec25", "peer": "+447700900000", "last_message": "hello", "last_timestamp": "2026-11-08T07:00:00Z", "unread_count": 2, "last_sms_id": 7}}})
		}
	})
	mux.HandleFunc("/api/sms/send", func(w http.ResponseWriter, r *http.Request) {
		if !authed(w, r) {
			return
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.sent = append(f.sent, body)
		f.mu.Unlock()
		write(w, 202, map[string]any{"data": map[string]any{"outcome": "accepted_unconfirmed"}})
	})
	mux.HandleFunc("/api/rotate", func(w http.ResponseWriter, r *http.Request) {
		// Changes the CSRF token, as a session refresh elsewhere does.
		setCSRF(w, "t2")
		f.mu.Lock()
		f.csrf = "t3" // the server now expects a token the client lacks
		f.rotated = true
		f.mu.Unlock()
		write(w, 200, map[string]any{"data": map[string]any{}})
	})
	mux.HandleFunc("/api/devices/ec25/flight-mode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("flight mode used %s", r.Method)
		}
		if authed(w, r) {
			write(w, 200, map[string]any{"data": map[string]any{}})
		}
	})
	return mux
}

func TestClientAgainstFakeHalo(t *testing.T) {
	f := &fakeHalo{}
	srv := httptest.NewServer(f.handler(t))
	defer srv.Close()
	c := NewClient()
	c.Configure(strings.TrimPrefix(srv.URL, "http://"), "")
	ctx := t.Context()

	if _, err := c.Login(ctx, "admin", "wrong"); err == nil || !strings.Contains(err.Error(), "用户名或密码错误") {
		t.Fatalf("bad password: %v", err)
	}
	s, err := c.Login(ctx, "admin", "secret")
	if err != nil || s.Name() != "admin" {
		t.Fatalf("login: %v %+v", err, s)
	}
	devs, err := c.Devices(ctx)
	if err != nil || len(devs) != 1 || !devs[0].LineReady() || devs[0].Modem.Operator != "Vodafone UK" || devs[0].Bars() != 3 {
		t.Fatalf("devices: %v %+v", err, devs)
	}
	ks, err := c.Contacts(ctx)
	if err != nil || len(ks) != 1 || ks[0].UnreadCount != 2 || ks[0].LastTime.IsZero() {
		t.Fatalf("contacts: %v %+v", err, ks)
	}
	if err := c.SendSMS(ctx, "ec25", "+447700900000", "hi"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if err := c.SetFlightMode(ctx, "ec25", true); err != nil {
		t.Fatalf("flight mode: %v", err)
	}
	// A stale CSRF token is refreshed from the session once, then retried.
	_ = c.do(ctx, http.MethodGet, "rotate", nil, nil)
	f.mu.Lock()
	f.csrf = "t2"
	f.mu.Unlock()
	c.mu.Lock()
	c.csrf = "stale"
	c.mu.Unlock()
	if err := c.SendSMS(ctx, "ec25", "+447700900000", "again"); err != nil {
		t.Fatalf("send after CSRF rotation: %v", err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sent) != 2 || f.sent[0]["device_id"] != "ec25" || f.sent[1]["message"] != "again" {
		t.Errorf("sent %+v", f.sent)
	}
}

func TestSignInFlowRemembersPassword(t *testing.T) {
	f := &fakeHalo{}
	srv := httptest.NewServer(f.handler(t))
	defer srv.Close()
	a, store := testApp(t)
	tt := ui.NewTester(a.view, 1120, 740)
	clickRole(t, tt, ui.RoleTextField, "服务器地址")
	tt.Type(srv.URL)
	clickRole(t, tt, ui.RoleTextField, "用户名")
	tt.Type("admin")
	clickRole(t, tt, ui.RoleTextField, "密码")
	tt.Type("secret")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.phase != phaseSignedIn || a.version != "1.1.15" {
		t.Fatalf("not signed in: phase %v err %q", a.phase, a.loginErr)
	}
	if store.Password(srv.URL, "admin") != "secret" {
		t.Error("password not kept in the credential store")
	}
	tt.Advance(time.Second)
	if !tt.HasText("EC25") || !tt.HasText("QDC507") {
		t.Errorf("devices not loaded: %q", tt.Texts())
	}
	// A new run signs in by itself.
	b := newApp(store)
	b.spawn = func(fn func()) { fn() }
	b.start()
	if b.phase != phaseSignedIn {
		t.Errorf("auto sign-in: phase %v err %q", b.phase, b.loginErr)
	}
}

func TestNormalizeServer(t *testing.T) {
	for in, want := range map[string]string{
		"nas:7575":                  "http://nas:7575",
		" https://halo.lan:7575/ ":  "https://halo.lan:7575",
		"http://10.0.0.2:7575/api":  "http://10.0.0.2:7575",
		"192.168.1.10:7575/#/login": "http://192.168.1.10:7575",
	} {
		if got := NormalizeServer(in); got != want {
			t.Errorf("NormalizeServer(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGroupHex(t *testing.T) {
	got := groupHex("a1b2c3d4e5f60718293a4b5c6d7e8f90")
	if got != "A1:B2:C3:D4:E5:F6:07:18\n29:3A:4B:5C:6D:7E:8F:90" {
		t.Errorf("groupHex = %q", got)
	}
}
