package main

import "github.com/egoist/mygo/ui"

// Line icons on a 24-unit grid, drawn for this app in the manner of SF
// Symbols' regular weight: round caps and joins, 1.7 strokes.
func icon(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

func solid(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">` + shapes + `</svg>`))
}

var (
	icToday    = icon(`<rect x="3.5" y="3.5" width="7" height="7" rx="2"/><rect x="13.5" y="3.5" width="7" height="7" rx="2"/><rect x="3.5" y="13.5" width="7" height="7" rx="2"/><rect x="13.5" y="13.5" width="7" height="7" rx="2"/>`)
	icMessages = icon(`<path d="M12 4c4.97 0 9 3.36 9 7.5S16.97 19 12 19c-1.1 0-2.15-.16-3.12-.46L4.5 20l1.2-3.4C4.03 15.27 3 13.47 3 11.5 3 7.36 7.03 4 12 4z"/>`)
	icPhone    = icon(`<path d="M6.6 3.5h2.6l1.4 4.1-2 1.5a12 12 0 0 0 6.3 6.3l1.5-2 4.1 1.4v2.6a2 2 0 0 1-2.2 2C10.5 18.9 5.1 13.5 4.6 5.7a2 2 0 0 1 2-2.2z"/>`)
	icDevices  = icon(`<rect x="6" y="3" width="12" height="18" rx="2.5"/><path d="M10 17.5h4"/><path d="M9.5 7.5v4M12 6.5v5M14.5 9v2.5"/>`)
	icNetwork  = icon(`<path d="M12 3.5 5 6.3v5.2c0 4.3 2.9 7.9 7 9 4.1-1.1 7-4.7 7-9V6.3z"/><path d="m9 12 2.2 2.2L15.5 10"/>`)
	icLogs     = icon(`<path d="M7 3.5h7.5L19 8v12.5H7a2 2 0 0 1-2-2v-13a2 2 0 0 1 2-2z"/><path d="M14 3.5V8h5"/><path d="M8.5 12.5h7M8.5 16h5"/>`)
	icSettings = icon(`<circle cx="12" cy="12" r="3"/><path d="M12 2.8v2.4M12 18.8v2.4M21.2 12h-2.4M5.2 12H2.8M18.5 5.5l-1.7 1.7M7.2 16.8l-1.7 1.7M18.5 18.5l-1.7-1.7M7.2 7.2 5.5 5.5"/><circle cx="12" cy="12" r="6.3"/>`)

	icSend      = solid(`<circle cx="12" cy="12" r="11"/><path d="M12 6.5 7.5 11l1.2 1.2 2.45-2.45V17.5h1.7V9.75l2.45 2.45L16.5 11z" fill="#fff"/>`)
	icDelete    = icon(`<path d="M9 5.5h10.5a1.5 1.5 0 0 1 1.5 1.5v10a1.5 1.5 0 0 1-1.5 1.5H9L3.5 12z"/><path d="m11.5 9.5 5 5M16.5 9.5l-5 5"/>`)
	icCall      = solid(`<path d="M6.6 3.5h2.6l1.4 4.1-2 1.5a12 12 0 0 0 6.3 6.3l1.5-2 4.1 1.4v2.6a2 2 0 0 1-2.2 2C10.5 18.9 5.1 13.5 4.6 5.7a2 2 0 0 1 2-2.2z"/>`)
	icHangup    = solid(`<path d="M12 9c3.9 0 7.4 1.2 9.3 3.1.5.5.6 1.3.2 1.9l-1.3 1.8a1.3 1.3 0 0 1-1.6.4l-2.9-1.2a1.3 1.3 0 0 1-.8-1.2v-1.6a14 14 0 0 0-5.8 0v1.6c0 .5-.3 1-.8 1.2L5.4 16.2a1.3 1.3 0 0 1-1.6-.4l-1.3-1.8a1.5 1.5 0 0 1 .2-1.9C4.6 10.2 8.1 9 12 9z"/>`)
	icIncoming  = icon(`<path d="M16.5 7.5 8 16M8 9.5V16h6.5"/>`)
	icOutgoing  = icon(`<path d="M8 16 16.5 7.5M10 7.5h6.5V14"/>`)
	icRefresh   = icon(`<path d="M19.5 12a7.5 7.5 0 1 1-2.2-5.3L19.5 9"/><path d="M19.5 4v5h-5"/>`)
	icPower     = icon(`<path d="M12 3.5v8"/><path d="M7.2 6.3a7.5 7.5 0 1 0 9.6 0"/>`)
	icAirplane  = icon(`<path d="M10.5 4.2c0-.9.7-1.7 1.5-1.7s1.5.8 1.5 1.7v5.3l7 4.2v2l-7-2v4.4l2.2 1.6v1.8L12 20.5l-3.7 1v-1.8l2.2-1.6v-4.4l-7 2v-2l7-4.2z"/>`)
	icWifi      = icon(`<path d="M2.8 9.2a13 13 0 0 1 18.4 0"/><path d="M5.9 12.4a8.6 8.6 0 0 1 12.2 0"/><path d="M9 15.6a4.2 4.2 0 0 1 6 0"/><circle cx="12" cy="18.7" r=".9" fill="currentColor"/>`)
	icSignOut   = icon(`<path d="M14 4.5H6.5a2 2 0 0 0-2 2v11a2 2 0 0 0 2 2H14"/><path d="M10 12h10M16.5 8.5 20 12l-3.5 3.5"/>`)
	icChevron   = icon(`<path d="m9.5 6 6 6-6 6"/>`)
	icCompose   = icon(`<path d="M12 4.5H6.5a2 2 0 0 0-2 2v11a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2V12"/><path d="M17.8 3.8a1.9 1.9 0 0 1 2.7 2.7l-7.3 7.3-3.4.7.7-3.4z"/>`)
	icShield    = icon(`<path d="M12 3.5 5 6.3v5.2c0 4.3 2.9 7.9 7 9 4.1-1.1 7-4.7 7-9V6.3z"/>`)
	icServer    = icon(`<rect x="3.5" y="4" width="17" height="7" rx="2"/><rect x="3.5" y="13" width="17" height="7" rx="2"/><path d="M7 7.5h.01M7 16.5h.01"/>`)
	icCPU       = icon(`<rect x="6" y="6" width="12" height="12" rx="2"/><rect x="9.5" y="9.5" width="5" height="5" rx="1"/><path d="M9 3v3M15 3v3M9 18v3M15 18v3M3 9h3M3 15h3M18 9h3M18 15h3"/>`)
	icMemory    = icon(`<rect x="3" y="7" width="18" height="10" rx="2"/><path d="M7 7v10M11 7v10M15 7v10M3 13h18"/>`)
	icDisk      = icon(`<ellipse cx="12" cy="6" rx="7.5" ry="2.8"/><path d="M4.5 6v12c0 1.5 3.4 2.8 7.5 2.8s7.5-1.3 7.5-2.8V6"/><path d="M4.5 12c0 1.5 3.4 2.8 7.5 2.8s7.5-1.3 7.5-2.8"/>`)
	icDown      = icon(`<path d="M12 4.5v15M6 13.5l6 6 6-6"/>`)
	icUp        = icon(`<path d="M12 19.5v-15M6 10.5l6-6 6 6"/>`)
	icClock     = icon(`<circle cx="12" cy="12" r="8.5"/><path d="M12 7.5V12l3 2"/>`)
	icBell      = icon(`<path d="M18 16.5H6l1.3-2V10a4.7 4.7 0 0 1 9.4 0v4.5z"/><path d="M10 19.5a2 2 0 0 0 4 0"/>`)
	icPaint     = icon(`<circle cx="12" cy="12" r="8.5"/><path d="M12 3.5a8.5 8.5 0 0 0 0 17z" fill="currentColor"/>`)
	icInfo      = icon(`<circle cx="12" cy="12" r="8.5"/><path d="M12 11v5.5"/><circle cx="12" cy="7.8" r=".9" fill="currentColor"/>`)
	icLock      = icon(`<rect x="5" y="10.5" width="14" height="10" rx="2"/><path d="M8.5 10.5V7.8a3.5 3.5 0 0 1 7 0v2.7"/>`)
	icHalo      = icon(`<circle cx="12" cy="12" r="8.2"/><circle cx="12" cy="12" r="3.6"/>`)
	icTunnel    = icon(`<path d="M3.5 20.5V12a8.5 8.5 0 0 1 17 0v8.5"/><path d="M8 20.5V12a4 4 0 0 1 8 0v8.5"/>`)
	icExternal  = icon(`<path d="M14 4.5h5.5V10M19.5 4.5 11 13"/><path d="M17 14v4a2 2 0 0 1-2 2H6.5a2 2 0 0 1-2-2V9.5a2 2 0 0 1 2-2H10"/>`)
	icWarn      = icon(`<path d="M12 4 21 19.5H3z"/><path d="M12 10v4.5"/><circle cx="12" cy="17" r=".8" fill="currentColor"/>`)
	icEmpty     = icon(`<path d="M3.5 13.5 6.5 5h11l3 8.5v5a2 2 0 0 1-2 2h-13a2 2 0 0 1-2-2z"/><path d="M3.5 13.5h5l1.5 2.5h4l1.5-2.5h5"/>`)
	icMinimize  = icon(`<path d="M6 12h12"/>`)
	icMaximize  = icon(`<rect x="6" y="6" width="12" height="12" rx="1.5"/>`)
	icRestore   = icon(`<rect x="6" y="9" width="9" height="9" rx="1.5"/><path d="M9 9V7.5A1.5 1.5 0 0 1 10.5 6h6A1.5 1.5 0 0 1 18 7.5v6a1.5 1.5 0 0 1-1.5 1.5H15"/>`)
	icClose     = icon(`<path d="m7 7 10 10M17 7 7 17"/>`)
	icKeypad    = solid(`<circle cx="6" cy="5" r="1.8"/><circle cx="12" cy="5" r="1.8"/><circle cx="18" cy="5" r="1.8"/><circle cx="6" cy="11" r="1.8"/><circle cx="12" cy="11" r="1.8"/><circle cx="18" cy="11" r="1.8"/><circle cx="6" cy="17" r="1.8"/><circle cx="12" cy="17" r="1.8"/><circle cx="18" cy="17" r="1.8"/><circle cx="12" cy="22" r="1.4"/>`)
	icSignalBar = solid(`<rect x="0" y="0" width="24" height="24" rx="3"/>`)
)
