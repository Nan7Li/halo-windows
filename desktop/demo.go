package main

import (
	"math"
	"time"
)

// The preview shows the app with made-up data, without a server.

func demoHost() DashboardHost {
	var h DashboardHost
	h.Host.CPUModel = "Intel Celeron J3160"
	h.Host.BoardModel = "QDC507"
	h.Perf.CPUPercent = 18
	h.Perf.MemoryPercent = 42
	h.Perf.MemoryUsedBytes = 3_400_000_000
	h.Perf.MemoryTotalBytes = 8_000_000_000
	h.Perf.DiskPercent = 61
	h.Perf.DiskUsedBytes = 72_000_000_000
	h.Perf.DiskTotalBytes = 118_000_000_000
	h.Perf.NetRxBps = 128_000
	h.Perf.NetTxBps = 54_000
	return h
}

func demoDevices() []Device {
	return []Device{
		{
			ID: "ec25", Name: "EC25 · 主卡", DeviceType: "quectel", Running: true, Healthy: true, ControlOnline: true,
			PublicIP: "114.247.12.8", SMSEnabled: true, VoWiFiEnabled: true, VoWiFiActive: true, RadioMode: "vowifi",
			NetworkConnected: true, Registration: "registered",
			VoWiFi: VoWiFiRuntime{Phase: "ready", SIMReady: true, AccessReady: true, TunnelReady: true, IMSReady: true, SMSReady: true, LocalPhone: "+447700900123"},
			Modem:  Modem{Operator: "Vodafone UK", NetworkMode: "LTE", RadioBand: "B20", SignalDbm: -87, SignalSinr: 14, IMEI: "860000000000001", ICCID: "8944100000000002201", Model: "EC25-EUX"},
		},
		{
			ID: "eg25", Name: "EG25 · 备卡", DeviceType: "quectel", Running: true, Healthy: true, ControlOnline: true,
			PublicIP: "106.13.44.2", SMSEnabled: true, RadioMode: "cellular", NetworkConnected: true, Registration: "registered",
			Modem: Modem{Operator: "中国联通", NetworkMode: "LTE", RadioBand: "B3", SignalDbm: -101, SignalSinr: 6, IMEI: "860000000000002", ICCID: "8986010000000008830", Model: "EG25-G"},
		},
		{
			ID: "stick", Name: "410 Stick", DeviceType: "native410", Running: true, Healthy: true, ControlOnline: true,
			SMSEnabled: true, FlightMode: true, RadioMode: "airplane", Registration: "unknown",
			Modem: Modem{Operator: "giffgaff", NetworkMode: "—", Model: "UFI 410"},
		},
	}
}

func demoTasks(now time.Time) []Task {
	return []Task{
		{ID: 1, Name: "夜间切换到备卡", Enabled: true, NextRunAt: now.Add(5*time.Hour + 12*time.Minute)},
		{ID: 2, Name: "每月保号短信", Enabled: true, NextRunAt: now.Add(9 * 24 * time.Hour)},
	}
}

func demoContacts(now time.Time) []Contact {
	return []Contact{
		{DeviceID: "ec25", DeviceName: "EC25 · 主卡", Peer: "Vodafone", LastMessage: "Your code is 482193. Don't share it with anyone.", LastTime: now.Add(-2 * time.Minute), Direction: "inbound", LastSMSID: 41, UnreadCount: 1},
		{DeviceID: "ec25", DeviceName: "EC25 · 主卡", Peer: "+12025550123", DisplayName: "Alex Chen", LastMessage: "Landed. Call you on WiFi Calling.", LastTime: now.Add(-3 * time.Hour), Direction: "inbound", LastSMSID: 40},
		{DeviceID: "eg25", DeviceName: "EG25 · 备卡", Peer: "10010", DisplayName: "中国联通", LastMessage: "您本月已使用流量 2.4 GB，剩余 17.6 GB。", LastTime: now.Add(-26 * time.Hour), Direction: "inbound", LastSMSID: 37},
		{DeviceID: "ec25", DeviceName: "EC25 · 主卡", Peer: "+447700900456", LastMessage: "好的，明天见", LastTime: now.Add(-4 * 24 * time.Hour), Direction: "outbound", LastSMSID: 30},
	}
}

func demoThread(k Contact, now time.Time) []Message {
	switch k.Peer {
	case "Vodafone":
		return []Message{
			{ID: 39, Direction: "inbound", Body: "Welcome to Vodafone. Your eSIM is active and WiFi Calling is ready.", Timestamp: now.Add(-50 * time.Hour)},
			{ID: 41, Direction: "inbound", Body: "Your code is 482193. Don't share it with anyone.", Timestamp: now.Add(-2 * time.Minute)},
		}
	case "+12025550123":
		return []Message{
			{ID: 34, Direction: "outbound", Body: "落地了吗？", Timestamp: now.Add(-5 * time.Hour), DeliveryState: "delivered"},
			{ID: 35, Direction: "inbound", Body: "还在滑行，等下打给你", Timestamp: now.Add(-4*time.Hour - 50*time.Minute)},
			{ID: 36, Direction: "outbound", Body: "好，我这边开着 WiFi Calling", Timestamp: now.Add(-4*time.Hour - 48*time.Minute), DeliveryState: "delivered"},
			{ID: 40, Direction: "inbound", Body: "Landed. Call you on WiFi Calling.", Timestamp: now.Add(-3 * time.Hour)},
		}
	}
	return []Message{{ID: 1, Direction: k.Direction, Body: k.LastMessage, Timestamp: k.LastTime}}
}

func demoCalls(now time.Time) []CallRecord {
	answered := now.Add(-2 * time.Hour)
	return []CallRecord{
		{ID: 9, DeviceID: "ec25", Number: "+12025550123", Direction: "incoming", State: "answered", StartedAt: now.Add(-2 * time.Hour), AnsweredAt: &answered, DurationSeconds: 252, Transport: "vowifi"},
		{ID: 8, DeviceID: "ec25", Number: "+447700900456", Direction: "incoming", State: "missed", StartedAt: now.Add(-20 * time.Hour), Transport: "vowifi"},
		{ID: 7, DeviceID: "eg25", Number: "10010", Direction: "outgoing", State: "ended", StartedAt: now.Add(-3 * 24 * time.Hour), DurationSeconds: 728, Transport: "cellular"},
	}
}

func demoTunnels() TunnelSnapshot {
	return TunnelSnapshot{Available: true, Tunnels: []Tunnel{
		{ID: "home", Name: "home-core", Interface: "wg0", Available: true, Running: true, Peers: 3, TransferRx: 1_820_000_000, TransferTx: 420_000_000, Endpoint: "vpn.example.net:51820", Autostart: true},
		{ID: "uk", Name: "exit-uk", Interface: "wg1", Available: true, Peers: 1, Endpoint: "uk-1.edge.example:51820"},
	}}
}

func demoLogs(now time.Time) []LogEntry {
	return []LogEntry{
		{Time: now.Add(-6 * time.Minute), Level: "info", Message: "IMS registered on Vodafone UK ePDG"},
		{Time: now.Add(-5 * time.Minute), Level: "info", Message: "SMS MO submitted via IMS, 1/1 parts accepted"},
		{Time: now.Add(-4 * time.Minute), Level: "warn", Message: "RSRP dropped to -112 dBm on EG25"},
		{Time: now.Add(-3 * time.Minute), Level: "debug", Message: "QMI WDS start-network cid=17"},
		{Time: now.Add(-1 * time.Minute), Level: "error", Message: "ePDG probe timed out after 5s (epdg.epc.mnc015.mcc234.pub.3gppnetwork.org)"},
	}
}

func demoNetHistory() [][2]float64 {
	out := make([][2]float64, 40)
	for i := range out {
		x := float64(i)
		out[i] = [2]float64{
			90_000 + 45_000*math.Sin(x/4) + 20_000*math.Sin(x/1.7),
			40_000 + 18_000*math.Sin(x/3+1) + 8_000*math.Cos(x/1.3),
		}
	}
	return out
}
