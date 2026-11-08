package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
)

// Prefs are the client's settings, kept as JSON in the app's user data
// folder. The password is not here: it is in the system's credential store
// (Keychain, Credential Manager, Secret Service), under the server and user.
type Prefs struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	// Pins maps a server to the SHA-256 of the self-signed certificate the
	// user trusted for it.
	Pins     map[string]string `json:"pins,omitempty"`
	Remember bool              `json:"remember"`

	Appearance string  `json:"appearance"` // "", "light", "dark"
	Accent     string  `json:"accent"`     // "" follows the system
	Clarity    float64 `json:"clarity"`    // the sidebar glass, 0 frosted to 1 clear
	Notify     bool    `json:"notify"`
	Page       string  `json:"page"`
}

func defaultPrefs() Prefs {
	return Prefs{Remember: true, Notify: true, Clarity: 0.35}
}

// prefsStore reads and writes Prefs and the password. Tests replace it.
type prefsStore interface {
	Load() Prefs
	Save(Prefs)
	Password(server, user string) string
	SetPassword(server, user, password string)
}

type diskPrefs struct{}

func (diskPrefs) path() string {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "prefs.json")
}

func (d diskPrefs) Load() Prefs {
	p := defaultPrefs()
	if path := d.path(); path != "" {
		if buf, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(buf, &p)
		}
	}
	return p
}

func (d diskPrefs) Save(p Prefs) {
	path := d.path()
	if path == "" {
		return
	}
	buf, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	tmp := path + ".tmp"
	if os.WriteFile(tmp, buf, 0o600) == nil {
		_ = os.Rename(tmp, path)
	}
}

func secretKey(server, user string) string {
	return strings.ToLower(NormalizeServer(server)) + "|" + user
}

func (diskPrefs) Password(server, user string) string {
	buf, err := mygo.Secrets.Get(secretKey(server, user))
	if err != nil {
		return ""
	}
	return string(buf)
}

func (diskPrefs) SetPassword(server, user, password string) {
	key := secretKey(server, user)
	if password == "" {
		_ = mygo.Secrets.Delete(key)
		return
	}
	_ = mygo.Secrets.Set(key, []byte(password))
}

// memPrefs keeps everything in memory, for tests and the preview.
type memPrefs struct {
	p         Prefs
	passwords map[string]string
}

func (m *memPrefs) Load() Prefs { return m.p }
func (m *memPrefs) Save(p Prefs) {
	m.p = p
}
func (m *memPrefs) Password(server, user string) string {
	return m.passwords[secretKey(server, user)]
}
func (m *memPrefs) SetPassword(server, user, password string) {
	if m.passwords == nil {
		m.passwords = map[string]string{}
	}
	m.passwords[secretKey(server, user)] = password
}
