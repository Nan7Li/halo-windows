package main

import (
	"context"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

const appVersion = "1.0.0"

// The pages of the sidebar.
const (
	pageToday    = "today"
	pageMessages = "messages"
	pagePhone    = "phone"
	pageDevices  = "devices"
	pageNetwork  = "network"
	pageLogs     = "logs"
	pageSettings = "settings"
)

// app is the client's state. The view reads and changes it on the main
// thread; work on the network runs on goroutines and hands its results
// back through update.
type app struct {
	win   *mygo.Window
	api   *Client
	store prefsStore
	prefs Prefs
	// update runs fn on the main thread and draws a frame; spawn starts
	// background work. Tests make both synchronous.
	update func(fn func())
	spawn  func(fn func())
	notify func(title, body, id string)
	now    func() time.Time

	// Session
	phase     phase
	user      string
	version   string
	loginErr  string
	busy      bool
	password  string
	untrusted *UntrustedCertError
	trustOpen bool

	page  string
	toast []string

	// Data
	host       DashboardHost
	devices    []Device
	tasks      []Task
	contacts   []Contact
	calls      []CallRecord
	live       []LiveCall
	tunnels    TunnelSnapshot
	logs       []LogEntry
	loaded     map[string]bool
	loadErr    map[string]string
	refreshing map[string]bool

	// Messages
	selected    string // Contact.Key
	thread      []Message
	threadList  ui.ListState
	threadKey   string
	draft       string
	sending     bool
	search      string
	composeOpen bool
	composeTo   string
	composeLine string

	// Phone
	dialed      string
	dialLine    string
	phoneTab    int // 0 all, 1 missed
	callBusy    bool
	callStarted map[string]time.Time

	// Devices
	deviceID    string
	rebootAsk   bool
	deviceBusy  map[string]bool
	logLevel    int
	logQuery    string
	logList     ui.ListState
	logFollow   bool
	signOutAsk  bool
	notifyReady bool
	seenSMS     map[string]int64
	seenCalls   map[string]bool

	mu      sync.Mutex
	stopped chan struct{}
}

type phase int

const (
	phaseStarting phase = iota
	phaseSignedOut
	phaseSignedIn
)

func newApp(store prefsStore) *app {
	a := &app{
		api:         NewClient(),
		store:       store,
		update:      func(fn func()) { fn() },
		spawn:       func(fn func()) { go fn() },
		notify:      func(string, string, string) {},
		now:         time.Now,
		loaded:      map[string]bool{},
		loadErr:     map[string]string{},
		refreshing:  map[string]bool{},
		deviceBusy:  map[string]bool{},
		callStarted: map[string]time.Time{},
		seenSMS:     map[string]int64{},
		seenCalls:   map[string]bool{},
		logFollow:   true,
		stopped:     make(chan struct{}),
	}
	a.prefs = store.Load()
	a.page = a.prefs.Page
	if a.page == "" {
		a.page = pageToday
	}
	a.logList.FollowEnd = true
	return a
}

// start signs in with the saved password, if there is one.
func (a *app) start() {
	p := a.prefs
	if p.Server == "" || p.Username == "" || !p.Remember {
		a.phase = phaseSignedOut
		return
	}
	a.phase = phaseStarting
	a.spawn(func() {
		pw := a.store.Password(p.Server, p.Username)
		if pw == "" {
			a.update(func() { a.phase = phaseSignedOut })
			return
		}
		a.signIn(p.Server, p.Username, pw)
	})
}

func (a *app) say(msg string) { a.toast = append(a.toast, msg) }

func errText(err error) string {
	var api *APIError
	if errors.As(err, &api) && api.Status == 401 {
		return "登录已过期，请重新登录"
	}
	return err.Error()
}

// signIn logs in on a goroutine and shows the app.
func (a *app) signIn(server, user, password string) {
	a.update(func() { a.busy, a.loginErr, a.untrusted = true, "", nil })
	a.api.Configure(server, a.prefs.Pins[strings.ToLower(NormalizeServer(server))])
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	s, err := a.api.Login(ctx, user, password)
	var version string
	if err == nil {
		if info, err := a.api.SystemInfo(ctx); err == nil {
			version = info.Version
		}
	}
	a.update(func() {
		a.busy = false
		if err != nil {
			var cert *UntrustedCertError
			if errors.As(err, &cert) {
				a.untrusted, a.trustOpen = cert, true
			}
			a.loginErr = errText(err)
			a.phase = phaseSignedOut
			return
		}
		a.user = s.Name()
		if a.user == "" {
			a.user = user
		}
		a.version = version
		a.prefs.Server, a.prefs.Username = NormalizeServer(server), user
		a.store.Save(a.prefs)
		a.password = ""
		a.enter()
	})
	if a.prefs.Remember {
		a.store.SetPassword(server, user, password)
	}
}

// trust pins the certificate the server showed and signs in again.
func (a *app) trust(server, user, password string) {
	if a.untrusted == nil {
		return
	}
	if a.prefs.Pins == nil {
		a.prefs.Pins = map[string]string{}
	}
	a.prefs.Pins[strings.ToLower(NormalizeServer(server))] = a.untrusted.Fingerprint
	a.store.Save(a.prefs)
	a.untrusted = nil
	a.spawn(func() { a.signIn(server, user, password) })
}

func (a *app) enterPreview() {
	a.api.SetPreview(true)
	a.user, a.version = "预览", "Halo · 预览数据"
	a.enter()
}

// enter shows the signed-in app and loads its first data.
func (a *app) enter() {
	a.phase = phaseSignedIn
	a.loaded, a.loadErr = map[string]bool{}, map[string]string{}
	a.notifyReady = false
	a.refreshAll()
}

func (a *app) signOut() {
	preview := a.api.Preview()
	a.phase = phaseSignedOut
	a.host, a.devices, a.tasks, a.contacts, a.calls, a.live = DashboardHost{}, nil, nil, nil, nil, nil
	a.tunnels, a.logs, a.thread, a.selected = TunnelSnapshot{}, nil, nil, ""
	a.api.SetPreview(false)
	if preview {
		return
	}
	server, user := a.prefs.Server, a.prefs.Username
	a.store.SetPassword(server, user, "")
	a.spawn(func() { a.api.Logout(context.Background()) })
}

// ---- Loading ----

// load runs fetch for a part of the data unless it is loading already.
func (a *app) load(key string, fetch func(ctx context.Context) (func(), error)) {
	if a.refreshing[key] {
		return
	}
	a.refreshing[key] = true
	a.spawn(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		apply, err := fetch(ctx)
		a.update(func() {
			a.refreshing[key] = false
			if a.phase != phaseSignedIn {
				return
			}
			if err != nil {
				var api *APIError
				if errors.As(err, &api) && api.Status == 401 && !a.api.Preview() {
					a.phase = phaseSignedOut
					a.loginErr = "登录已过期，请重新登录"
					return
				}
				a.loadErr[key] = errText(err)
				return
			}
			delete(a.loadErr, key)
			a.loaded[key] = true
			if apply != nil {
				apply()
			}
		})
	})
}

func (a *app) refreshAll() {
	a.loadHost()
	a.loadDevices()
	a.loadTasks()
	a.loadContacts()
	a.loadCalls()
	a.loadTunnels()
	a.loadLogs()
}

func (a *app) loadHost() {
	a.load("host", func(ctx context.Context) (func(), error) {
		if a.api.Preview() {
			return func() { a.host = demoHost() }, nil
		}
		h, err := a.api.Host(ctx)
		return func() { a.host = h }, err
	})
}

func (a *app) loadDevices() {
	a.load("devices", func(ctx context.Context) (func(), error) {
		if a.api.Preview() {
			return func() { a.devices = demoDevices() }, nil
		}
		d, err := a.api.Devices(ctx)
		return func() { a.devices = d }, err
	})
}

func (a *app) loadTasks() {
	a.load("tasks", func(ctx context.Context) (func(), error) {
		if a.api.Preview() {
			return func() { a.tasks = demoTasks(a.now()) }, nil
		}
		t, err := a.api.Tasks(ctx)
		var api *APIError
		if errors.As(err, &api) && api.Status == 404 {
			err = nil // older Halo without automatic tasks
		}
		return func() { a.tasks = t }, err
	})
}

func (a *app) loadContacts() {
	a.load("contacts", func(ctx context.Context) (func(), error) {
		var list []Contact
		var err error
		if a.api.Preview() {
			list = demoContacts(a.now())
		} else {
			list, err = a.api.Contacts(ctx)
		}
		sort.SliceStable(list, func(i, j int) bool { return list[i].LastTime.After(list[j].LastTime) })
		return func() {
			a.contacts = list
			a.notifyMessages()
		}, err
	})
}

func (a *app) loadCalls() {
	a.load("calls", func(ctx context.Context) (func(), error) {
		if a.api.Preview() {
			return func() { a.calls = demoCalls(a.now()) }, nil
		}
		c, err := a.api.CallHistory(ctx)
		return func() { a.calls = c }, err
	})
}

func (a *app) loadTunnels() {
	a.load("tunnels", func(ctx context.Context) (func(), error) {
		if a.api.Preview() {
			return func() { a.tunnels = demoTunnels() }, nil
		}
		t, err := a.api.Tunnels(ctx)
		return func() { a.tunnels = t }, err
	})
}

var logLevels = []string{"", "info", "warn", "error"}

func (a *app) loadLogs() {
	level := logLevels[a.logLevel]
	a.load("logs", func(ctx context.Context) (func(), error) {
		if a.api.Preview() {
			return func() { a.logs = filterLogs(demoLogs(a.now()), level) }, nil
		}
		l, err := a.api.Logs(ctx, level)
		return func() { a.logs = l }, err
	})
}

func filterLogs(in []LogEntry, level string) []LogEntry {
	if level == "" {
		return in
	}
	rank := map[string]int{"debug": 0, "info": 1, "warn": 2, "warning": 2, "error": 3}
	var out []LogEntry
	for _, e := range in {
		if rank[strings.ToLower(e.Level)] >= rank[level] {
			out = append(out, e)
		}
	}
	return out
}

// loadLive polls the calls in progress on every line that can take calls.
func (a *app) loadLive() {
	if a.api.Preview() {
		return
	}
	var ids []string
	for _, d := range a.devices {
		if d.LineReady() {
			ids = append(ids, d.ID)
		}
	}
	if len(ids) == 0 {
		if len(a.live) > 0 {
			a.live = nil
		}
		return
	}
	a.load("live", func(ctx context.Context) (func(), error) {
		var all []LiveCall
		for _, id := range ids {
			calls, err := a.api.LiveCalls(ctx, id)
			if err != nil {
				continue
			}
			all = append(all, calls...)
		}
		return func() { a.setLive(all) }, nil
	})
}

// setLive takes the calls in progress, notifying of new incoming ones and
// reloading the history when one ends.
func (a *app) setLive(all []LiveCall) {
	var live []LiveCall
	ended := false
	for _, l := range all {
		if l.Phase() == PhaseEnded {
			continue
		}
		live = append(live, l)
	}
	for _, old := range a.live {
		still := false
		for _, l := range live {
			if l.ID == old.ID {
				still = true
			}
		}
		if !still {
			ended = true
		}
	}
	for _, l := range live {
		if l.Phase() == PhaseIncoming && !a.seenCalls[l.ID] {
			a.seenCalls[l.ID] = true
			a.notify("来电", displayNumber(l.Number), "call/"+l.ID)
		}
	}
	a.live = live
	if ended {
		a.loadCalls()
	}
}

// notifyMessages notifies of messages that came in since the last load,
// once the first load has set the baseline.
func (a *app) notifyMessages() {
	for _, k := range a.contacts {
		key := k.Key()
		last, seen := a.seenSMS[key]
		if a.notifyReady && a.prefs.Notify && (!seen || k.LastSMSID > last) &&
			k.UnreadCount > 0 && k.Direction != "outbound" && key != a.selected {
			a.notify(k.Title(), k.LastMessage, "sms/"+key)
		}
		a.seenSMS[key] = k.LastSMSID
	}
	a.notifyReady = true
}

// poll refreshes what the page shows, and the calls, while signed in.
func (a *app) poll() {
	tick := 0
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-a.stopped:
			return
		case <-t.C:
		}
		tick++
		a.update(func() {
			if a.phase != phaseSignedIn || a.api.Preview() {
				return
			}
			a.loadLive()
			if tick%5 == 0 {
				a.loadContacts()
				a.loadDevices()
			}
			switch a.page {
			case pageToday:
				if tick%2 == 0 {
					a.loadHost()
				}
			case pageMessages:
				if tick%3 == 0 && a.selected != "" {
					a.openThread(a.selected, false)
				}
			case pagePhone:
				if tick%10 == 0 {
					a.loadCalls()
				}
			case pageNetwork:
				if tick%5 == 0 {
					a.loadTunnels()
				}
			case pageLogs:
				if tick%3 == 0 {
					a.loadLogs()
				}
			}
		})
	}
}

// ---- Actions ----

// act runs a write on a goroutine, then shows its outcome and reloads.
func (a *app) act(busy *bool, done string, work func(ctx context.Context) error, after func()) {
	if a.api.Preview() {
		a.say(done + "（预览，未发送到服务器）")
		if after != nil {
			after()
		}
		return
	}
	if busy != nil {
		*busy = true
	}
	a.spawn(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := work(ctx)
		a.update(func() {
			if busy != nil {
				*busy = false
			}
			if err != nil {
				a.say(errText(err))
				return
			}
			if done != "" {
				a.say(done)
			}
			if after != nil {
				after()
			}
		})
	})
}

func (a *app) contactByKey(key string) (Contact, bool) {
	for _, k := range a.contacts {
		if k.Key() == key {
			return k, true
		}
	}
	return Contact{}, false
}

// openThread selects a conversation and loads its messages; Halo marks
// them read as it serves them.
func (a *app) openThread(key string, fresh bool) {
	k, ok := a.contactByKey(key)
	if !ok {
		return
	}
	if fresh {
		if a.selected != key {
			a.threadList = ui.ListState{FollowEnd: true}
		}
		a.selected, a.draft = key, ""
		if a.threadKey != key {
			a.thread = nil
		}
	}
	a.load("thread", func(ctx context.Context) (func(), error) {
		var msgs []Message
		var err error
		if a.api.Preview() {
			msgs = demoThread(k, a.now())
		} else {
			msgs, err = a.api.Thread(ctx, k)
		}
		return func() {
			if a.selected != key {
				return
			}
			a.thread, a.threadKey = msgs, key
			for i := range a.contacts {
				if a.contacts[i].Key() == key {
					a.contacts[i].UnreadCount = 0
				}
			}
		}, err
	})
}

func (a *app) sendMessage() {
	k, ok := a.contactByKey(a.selected)
	text := strings.TrimSpace(a.draft)
	if !ok || text == "" || a.sending {
		return
	}
	a.draft = ""
	pending := Message{ID: -a.now().UnixNano(), Direction: "outbound", Body: text, Timestamp: a.now(), DeliveryState: "sending"}
	a.thread = append(a.thread, pending)
	key := a.selected
	if a.api.Preview() {
		a.thread[len(a.thread)-1].DeliveryState = "delivered"
		return
	}
	a.act(&a.sending, "", func(ctx context.Context) error {
		return a.api.SendSMS(ctx, k.DeviceID, k.Peer, text)
	}, func() {
		if a.selected == key {
			a.openThread(key, false)
		}
		a.loadContacts()
	})
}

func (a *app) sendNew() {
	to := strings.TrimSpace(a.composeTo)
	text := strings.TrimSpace(a.draft)
	if to == "" || text == "" {
		return
	}
	line := a.composeLine
	a.act(&a.sending, "短信已发送", func(ctx context.Context) error {
		return a.api.SendSMS(ctx, line, to, text)
	}, func() {
		a.composeOpen, a.draft, a.composeTo = false, "", ""
		a.loadContacts()
	})
}

func (a *app) dial(number string) {
	number = strings.TrimSpace(number)
	if number == "" {
		return
	}
	line := a.dialLine
	if line == "" {
		if ready := a.readyLines(); len(ready) > 0 {
			line = ready[0].ID
		}
	}
	if line == "" {
		a.say("没有可以通话的线路：需要开启 WiFi 通话并完成 IMS 注册")
		return
	}
	a.act(&a.callBusy, "正在呼叫 "+displayNumber(number), func(ctx context.Context) error {
		return a.api.Dial(ctx, line, number)
	}, func() { a.loadLive() })
}

func (a *app) answer(l LiveCall) {
	a.act(&a.callBusy, "", func(ctx context.Context) error { return a.api.Answer(ctx, l.DeviceID, l.ID) }, func() { a.loadLive() })
}

func (a *app) hangup(l LiveCall) {
	a.act(&a.callBusy, "", func(ctx context.Context) error { return a.api.Hangup(ctx, l.DeviceID, l.ID) }, func() {
		a.loadLive()
		a.loadCalls()
	})
}

func (a *app) readyLines() []Device {
	var out []Device
	for _, d := range a.devices {
		if d.LineReady() {
			out = append(out, d)
		}
	}
	return out
}

func displayNumber(n string) string {
	if strings.TrimSpace(n) == "" {
		return "未知号码"
	}
	return n
}

// ---- Main ----

func main() {
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	a := newApp(diskPrefs{})
	mygo.App.WhenReady(func() {
		a.win = mygo.NewWindow(mygo.WindowOptions{
			Title:          "Halo",
			Width:          1120,
			Height:         740,
			MinWidth:       860,
			MinHeight:      580,
			StateKey:       "main",
			TitleBarStyle:  mygo.TitleBarHiddenInset,
			TitleBarHeight: titleBarHeight,
			Content:        ui.View(a.view),
		})
		win := a.win
		a.update = func(fn func()) { win.Update(fn) }
		a.notify = func(title, body, id string) {
			n := mygo.NewNotification(mygo.NotificationOptions{ID: id, Title: title, Body: body})
			go func() {
				if err := n.Show(); err != nil {
					log.Print(err)
				}
			}()
		}
		a.start()
		go a.poll()
	})
	mygo.App.OnSecondInstance(func([]string, string) {
		if a.win != nil {
			a.win.Show()
			a.win.Focus()
		}
	})
	mygo.App.OnNotificationClick(func(id string) {
		if a.win == nil {
			return
		}
		a.win.Update(func() {
			switch {
			case strings.HasPrefix(id, "sms/"):
				a.page = pageMessages
				a.openThread(strings.TrimPrefix(id, "sms/"), true)
			case strings.HasPrefix(id, "call/"):
				a.page = pagePhone
			}
		})
		a.win.Show()
		a.win.Focus()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
	close(a.stopped)
}
