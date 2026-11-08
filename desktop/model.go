package main

import (
	"fmt"
	"strings"
	"time"
)

type DashboardHost struct {
	Host struct {
		CPUModel   string `json:"cpu_model"`
		BoardModel string `json:"board_model"`
	} `json:"host"`
	Perf struct {
		CPUPercent       float64 `json:"cpu_percent"`
		MemoryPercent    float64 `json:"memory_percent"`
		MemoryUsedBytes  int64   `json:"memory_used_bytes"`
		MemoryTotalBytes int64   `json:"memory_total_bytes"`
		DiskPercent      float64 `json:"disk_percent"`
		DiskUsedBytes    int64   `json:"disk_used_bytes"`
		DiskTotalBytes   int64   `json:"disk_total_bytes"`
		NetRxBps         float64 `json:"net_rx_bps"`
		NetTxBps         float64 `json:"net_tx_bps"`
	} `json:"perf"`
}

type Modem struct {
	Operator    string  `json:"operator"`
	NetworkMode string  `json:"network_mode"`
	RadioBand   string  `json:"radio_band"`
	SignalDbm   float64 `json:"signal_dbm"`
	SignalSinr  float64 `json:"signal_sinr"`
	IMEI        string  `json:"imei"`
	ICCID       string  `json:"iccid"`
	IMSI        string  `json:"imsi"`
	Model       string  `json:"model"`
	Firmware    string  `json:"firmware"`
}

type VoWiFiRuntime struct {
	Phase       string `json:"phase"`
	SIMReady    bool   `json:"sim_ready"`
	AccessReady bool   `json:"access_ready"`
	TunnelReady bool   `json:"tunnel_ready"`
	IMSReady    bool   `json:"ims_ready"`
	SMSReady    bool   `json:"sms_ready"`
	LocalPhone  string `json:"local_phone"`
	LastError   string `json:"last_error"`
}

type Device struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	DeviceType         string        `json:"device_type"`
	Running            bool          `json:"running"`
	Healthy            bool          `json:"healthy"`
	ControlOnline      bool          `json:"control_online"`
	PublicIP           string        `json:"public_ip"`
	SMSEnabled         bool          `json:"sms_enabled"`
	VoWiFiEnabled      bool          `json:"vowifi_enabled"`
	VoWiFiActive       bool          `json:"vowifi_active"`
	CallAudioReady     bool          `json:"call_audio_ready"`
	CallAudioTransport string        `json:"call_audio_transport"`
	RadioMode          string        `json:"radio_mode"`
	FlightMode         bool          `json:"flight_mode"`
	NetworkConnected   bool          `json:"network_connected"`
	ModemPhase         string        `json:"modem_phase"`
	Registration       string        `json:"registration_state_label"`
	VoWiFi             VoWiFiRuntime `json:"vowifi_runtime"`
	Modem              Modem         `json:"modem"`
}

// LineReady tells whether calls can be placed on the device, as the web
// UI's lineReady does.
func (d Device) LineReady() bool {
	if d.CallAudioTransport == "cellular" || d.CallAudioTransport == "vowifi" {
		return d.CallAudioReady
	}
	return d.VoWiFiEnabled && d.VoWiFi.IMSReady
}

func (d Device) Title() string {
	if d.Name != "" {
		return d.Name
	}
	return d.ID
}

// Status is a short state for lists, and whether it is good.
func (d Device) Status() (string, Tone) {
	switch {
	case d.ModemPhase == "rebooting":
		return "正在重启", ToneWarning
	case !d.Running:
		return "离线", ToneDanger
	case d.FlightMode || d.RadioMode == "airplane":
		return "飞行模式", ToneWarning
	case d.VoWiFiEnabled && d.VoWiFi.IMSReady:
		return "WiFi 通话就绪", ToneGood
	case d.Registration == "searching":
		return "正在搜索网络", ToneWarning
	case d.Registration == "denied":
		return "注册被拒绝", ToneDanger
	case d.Healthy:
		return "在线", ToneGood
	}
	return "状态未知", ToneMuted
}

// Bars is the signal strength as 0–4 bars.
func (d Device) Bars() int {
	s := d.Modem.SignalDbm
	switch {
	case s == 0:
		return 0
	case s >= -85:
		return 4
	case s >= -95:
		return 3
	case s >= -105:
		return 2
	case s >= -115:
		return 1
	}
	return 0
}

type Tone int

const (
	ToneMuted Tone = iota
	ToneGood
	ToneWarning
	ToneDanger
)

type Task struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	NextRunAt time.Time `json:"next_run_at"`
}

type Contact struct {
	DeviceID     string    `json:"device_id"`
	DeviceName   string    `json:"device_name"`
	ModemIMEI    string    `json:"modem_imei"`
	ICCID        string    `json:"iccid"`
	IMSI         string    `json:"imsi"`
	LocalPhone   string    `json:"local_phone"`
	Peer         string    `json:"peer"`
	DisplayName  string    `json:"display_name"`
	LastMessage  string    `json:"last_message"`
	LastTime     time.Time `json:"last_timestamp"`
	Direction    string    `json:"direction"`
	LastSMSID    int64     `json:"last_sms_id"`
	UnreadCount  int       `json:"unread_count"`
	MessageCount int       `json:"message_count"`
}

// Key identifies a conversation: one peer on one line.
func (k Contact) Key() string {
	return k.DeviceID + "|" + k.ModemIMEI + "|" + k.IMSI + "|" + k.Peer
}

func (k Contact) Title() string {
	if strings.TrimSpace(k.DisplayName) != "" {
		return k.DisplayName
	}
	return k.Peer
}

type Message struct {
	ID            int64     `json:"id"`
	DeviceID      string    `json:"device_id"`
	Peer          string    `json:"peer"`
	Direction     string    `json:"direction"`
	Body          string    `json:"body"`
	Content       string    `json:"content"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
	DeliveryState string    `json:"delivery_state"`
}

func (m Message) Mine() bool { return m.Direction == "outbound" || m.Direction == "sent" }

func (m Message) Text() string {
	if m.Body != "" {
		return m.Body
	}
	return m.Content
}

type CallRecord struct {
	ID              int64      `json:"id"`
	DeviceID        string     `json:"device_id"`
	Number          string     `json:"number"`
	Direction       string     `json:"direction"`
	State           string     `json:"state"`
	StartedAt       time.Time  `json:"started_at"`
	AnsweredAt      *time.Time `json:"answered_at"`
	DurationSeconds int        `json:"duration_seconds"`
	Transport       string     `json:"transport"`
}

func (r CallRecord) Missed() bool {
	return r.Direction == "incoming" && (r.State == "missed" || r.AnsweredAt == nil && r.State != "active")
}

// LiveCall is a call in progress on a line, from /devices/{id}/calls.
type LiveCall struct {
	DeviceID   string     `json:"-"`
	ID         string     `json:"id"`
	Number     string     `json:"number"`
	Direction  string     `json:"direction"`
	State      string     `json:"state"`
	StartedAt  time.Time  `json:"started_at"`
	AnsweredAt *time.Time `json:"answered_at"`
}

type CallPhase int

const (
	PhaseDialing CallPhase = iota
	PhaseIncoming
	PhaseConnecting
	PhaseActive
	PhaseEnded
)

func (l LiveCall) Phase() CallPhase {
	switch {
	case l.State == "ended" || l.State == "failed":
		return PhaseEnded
	case l.Direction == "incoming" && l.State == "ringing":
		return PhaseIncoming
	case l.State == "active":
		return PhaseActive
	case l.State == "early_media":
		return PhaseConnecting
	}
	return PhaseDialing
}

type Tunnel struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Interface       string `json:"interface"`
	Autostart       bool   `json:"autostart"`
	Available       bool   `json:"available"`
	Running         bool   `json:"running"`
	Peers           int    `json:"peers"`
	TransferRx      int64  `json:"transfer_rx"`
	TransferTx      int64  `json:"transfer_tx"`
	Endpoint        string `json:"endpoint"`
	LatestHandshake string `json:"latest_handshake"`
	Error           string `json:"error"`
	External        bool   `json:"external"`
}

type TunnelSnapshot struct {
	Available bool     `json:"available"`
	Hint      string   `json:"hint"`
	Tunnels   []Tunnel `json:"tunnels"`
}

type LogEntry struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"fields"`
}

// ---- Formatting ----

func formatBytes(n float64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%.0f B", n)
	case n < 1<<20:
		return fmt.Sprintf("%.1f KB", n/1024)
	case n < 1<<30:
		return fmt.Sprintf("%.1f MB", n/(1<<20))
	}
	return fmt.Sprintf("%.2f GB", n/(1<<30))
}

func formatRate(bps float64) string { return formatBytes(bps) + "/s" }

// relTime is a time as the Messages app lists it: the clock today,
// "昨天", the weekday this week, else the date.
func relTime(t, now time.Time) string {
	if t.IsZero() || t.Year() < 2000 {
		return ""
	}
	t = t.Local()
	now = now.Local()
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	today := time.Date(y2, m2, d2, 0, 0, 0, 0, time.Local)
	day := time.Date(y1, m1, d1, 0, 0, 0, 0, time.Local)
	days := int(today.Sub(day).Hours() / 24)
	switch {
	case days <= 0:
		return t.Format("15:04")
	case days == 1:
		return "昨天"
	case days < 7:
		return [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}[t.Weekday()]
	case y1 == y2:
		return fmt.Sprintf("%d月%d日", m1, d1)
	}
	return fmt.Sprintf("%d/%d/%d", y1, m1, d1)
}

// untilTime is how long until t, for the next run of a task.
func untilTime(t, now time.Time) string {
	if t.IsZero() || t.Year() < 2000 {
		return "—"
	}
	d := t.Sub(now)
	switch {
	case d < 0:
		return "即将执行"
	case d < time.Hour:
		return fmt.Sprintf("%d 分钟后", int(d.Minutes())+1)
	case d < 24*time.Hour:
		return t.Local().Format("今天 15:04")
	}
	return t.Local().Format("1月2日 15:04")
}

func formatDuration(sec int) string {
	if sec <= 0 {
		return ""
	}
	if sec >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", sec/3600, sec/60%60, sec%60)
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

// initials is what an avatar shows for a name or a number.
func initials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "#"
	}
	r := []rune(name)
	if r[0] == '+' || r[0] >= '0' && r[0] <= '9' {
		return "#"
	}
	if r[0] < 128 {
		parts := strings.Fields(name)
		s := strings.ToUpper(string([]rune(parts[0])[:1]))
		if len(parts) > 1 {
			s += strings.ToUpper(string([]rune(parts[len(parts)-1])[:1]))
		}
		return s
	}
	return string(r[len(r)-1:])
}
