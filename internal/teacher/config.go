// Package teacher is the herdr-teacher wrapper over the soho engine.
//
// soho is the transport (spawn, dispatch, wait, lanes, family check,
// roster). This package only:
//  1. points the engine at this skill's roles (<prefix>ROLES);
//  2. layers the teacher config (skill defaults < user < project) and
//     exports each key as <prefix><KEY>, unless the environment already
//     sets it (an explicit env var or flag still wins);
//  3. enforces the segregation rules the engine cannot express on its own;
//  4. runs the engine with the arguments unchanged.
package teacher

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Env is a process environment as a map. Lookups are exact-case.
type Env map[string]string

// Get returns the value of name, or "" when it is unset.
func (e Env) Get(name string) string { return e[name] }

// List returns the environment as KEY=VALUE pairs for exec.
func (e Env) List() []string {
	out := make([]string, 0, len(e))
	for k, v := range e {
		out = append(out, k+"="+v)
	}
	return out
}

// EnvFromOS reads the current process environment.
func EnvFromOS() Env {
	env := Env{}
	for _, kv := range os.Environ() {
		if i := strings.IndexByte(kv, '='); i > 0 {
			env[kv[:i]] = kv[i+1:]
		}
	}
	return env
}

var (
	spaceRE   = regexp.MustCompile(`\s+`)
	invalidRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)
)

// NormalizeKey is the soho key normalization: no whitespace, `.` and `-`
// become `_`, anything else outside [A-Za-z0-9_] is dropped.
func NormalizeKey(raw string) string {
	s := spaceRE.ReplaceAllString(raw, "")
	s = strings.NewReplacer(".", "_", "-", "_").Replace(s)
	return invalidRE.ReplaceAllString(s, "")
}

// KV is one key=value line of a config file.
type KV struct{ Key, Value string }

// ParseConfig reads key=value lines; `#` starts a comment line, blank lines
// and lines without a key are skipped.
func ParseConfig(text string) []KV {
	var out []KV
	text = strings.ReplaceAll(text, "\r\n", "\n")
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		out = append(out, KV{strings.TrimSpace(line[:eq]), strings.TrimSpace(line[eq+1:])})
	}
	return out
}

// Entry is one effective config key and the layer it came from.
type Entry struct {
	Key    string // as written in the file
	Value  string
	Source string // defaults | user | project
	File   string
}

// Config is the effective teacher config, in first-seen key order.
type Config struct {
	order   []string
	entries map[string]Entry
}

// Get returns the entry for a normalized key.
func (c *Config) Get(normalized string) (Entry, bool) {
	e, ok := c.entries[normalized]
	return e, ok
}

// Value returns the value of a normalized key, or fallback.
func (c *Config) Value(normalized, fallback string) string {
	if e, ok := c.entries[normalized]; ok && e.Value != "" {
		return e.Value
	}
	return fallback
}

// Each calls fn for every effective key, in first-seen order.
func (c *Config) Each(fn func(normalized string, e Entry)) {
	for _, nk := range c.order {
		fn(nk, c.entries[nk])
	}
}

func (c *Config) set(nk string, e Entry) {
	if _, ok := c.entries[nk]; !ok {
		c.order = append(c.order, nk)
	}
	c.entries[nk] = e
}

// UserConfigPath is the user layer: $XDG_CONFIG_HOME/herdr-teacher/config,
// else %APPDATA%\herdr-teacher\config on Windows, else
// ~/.config/herdr-teacher/config.
func UserConfigPath(env Env, goos string) string {
	if x := env.Get("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "herdr-teacher", "config")
	}
	if goos == "windows" {
		base := env.Get("APPDATA")
		if base == "" {
			base = filepath.Join(homeDir(env, goos), "AppData", "Roaming")
		}
		return filepath.Join(base, "herdr-teacher", "config")
	}
	return filepath.Join(homeDir(env, goos), ".config", "herdr-teacher", "config")
}

func homeDir(env Env, goos string) string {
	if goos == "windows" {
		if h := env.Get("USERPROFILE"); h != "" {
			return h
		}
	}
	if h := env.Get("HOME"); h != "" {
		return h
	}
	h, _ := os.UserHomeDir()
	return h
}

// ProjectRoot is the git toplevel of cwd, or cwd itself.
func ProjectRoot(cwd string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err == nil {
		if p := strings.TrimSpace(string(out)); p != "" {
			return filepath.FromSlash(p)
		}
	}
	return cwd
}

// ProjectConfigPath is the project layer for a project root.
func ProjectConfigPath(root string) string {
	return filepath.Join(root, ".agents", "herdr-teacher.conf")
}

// LoadConfig layers <skill>/config.defaults, the user file and the project
// file; a later layer replaces a key's value and source.
func LoadConfig(skillDir string, env Env, cwd string) *Config {
	cfg := &Config{entries: map[string]Entry{}}
	layers := []struct{ source, file string }{
		{"defaults", filepath.Join(skillDir, "config.defaults")},
		{"user", UserConfigPath(env, runtime.GOOS)},
		{"project", ProjectConfigPath(ProjectRoot(cwd))},
	}
	for _, l := range layers {
		data, err := os.ReadFile(l.file)
		if err != nil {
			continue
		}
		for _, kv := range ParseConfig(string(data)) {
			cfg.set(NormalizeKey(kv.Key), Entry{Key: kv.Key, Value: kv.Value, Source: l.source, File: l.file})
		}
	}
	return cfg
}
