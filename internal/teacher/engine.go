package teacher

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Engine is the soho build that runs the team.
type Engine struct {
	// Name says which build was found: "soho" (the Go binary, SOHO_*),
	// "herdr-soho" (the former Go binary, HERDR_SOHO_*) or "herdr-soho.mjs"
	// (the former JavaScript engine, HERDR_SOHO_*).
	Name string
	// Prefix of every engine environment variable.
	Prefix string
	// Program and Args run the engine; the user's arguments go after Args.
	Program string
	Args    []string
	// SkillDir is the herdr-soho skill tree the JavaScript engine needs
	// (exported as HERDR_SOHO_SKILL_DIR); empty for the Go binaries, which
	// find or bundle their own.
	SkillDir string
}

// Command returns the program and its full argument list.
func (e Engine) Command(user []string) (string, []string) {
	return e.Program, append(append([]string{}, e.Args...), user...)
}

// LookPath finds an executable, like exec.LookPath.
type LookPath func(file string) (string, error)

// ErrNoEngine means neither soho nor the former herdr-soho was found.
var ErrNoEngine = errors.New("no soho engine found: install soho (https://github.com/djalmajr/soho), or set HERDR_TEACHER_SOHO_BIN, or keep the former herdr-soho skill next to this one")

// FindEngine resolves the engine, first match wins:
//  1. HERDR_TEACHER_SOHO_BIN — a soho Go binary (SOHO_*);
//  2. `soho` on PATH (SOHO_*);
//  3. HERDR_SOHO_BIN — the former herdr-soho Go binary (HERDR_SOHO_*);
//  4. the former herdr-soho skill (HERDR_TEACHER_SOHO_DIR, the sibling
//     ../herdr-soho, ~/.agents/skills/herdr-soho, ~/.claude/skills/herdr-soho)
//     run with node, else bun (HERDR_SOHO_*).
func FindEngine(env Env, teacherSkillDir string, look LookPath) (Engine, error) {
	if bin := env.Get("HERDR_TEACHER_SOHO_BIN"); bin != "" {
		if !isFile(bin) {
			return Engine{}, errors.New("HERDR_TEACHER_SOHO_BIN is not a file: " + bin)
		}
		return Engine{Name: "soho", Prefix: "SOHO_", Program: bin}, nil
	}
	if p, err := look("soho"); err == nil {
		return Engine{Name: "soho", Prefix: "SOHO_", Program: p}, nil
	}
	if bin := env.Get("HERDR_SOHO_BIN"); bin != "" && isFile(bin) {
		return Engine{Name: "herdr-soho", Prefix: "HERDR_SOHO_", Program: bin}, nil
	}
	home := homeDir(env, runtime.GOOS)
	candidates := []string{
		env.Get("HERDR_TEACHER_SOHO_DIR"),
		filepath.Join(filepath.Dir(teacherSkillDir), "herdr-soho"),
		filepath.Join(home, ".agents", "skills", "herdr-soho"),
		filepath.Join(home, ".claude", "skills", "herdr-soho"),
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		real, err := filepath.EvalSymlinks(c)
		if err != nil {
			continue
		}
		mjs := filepath.Join(real, "scripts", "herdr-soho.mjs")
		if !isFile(mjs) {
			continue
		}
		for _, runtimeName := range []string{"node", "bun"} {
			if p, err := look(runtimeName); err == nil {
				return Engine{Name: "herdr-soho.mjs", Prefix: "HERDR_SOHO_", Program: p, Args: []string{mjs}, SkillDir: real}, nil
			}
		}
		return Engine{}, errors.New("found the herdr-soho skill at " + real + " but neither node nor bun is on PATH")
	}
	return Engine{}, ErrNoEngine
}

// EngineEnv is the environment the engine runs with, and the names it
// exported from the teacher config. `teacher.*` keys belong to this wrapper
// and the professor; they are never exported.
func EngineEnv(cfg *Config, env Env, eng Engine, teacherSkillDir string) (Env, []string) {
	out := Env{}
	for k, v := range env {
		out[k] = v
	}
	var exported []string
	cfg.Each(func(nk string, e Entry) {
		if strings.HasPrefix(nk, "teacher_") || e.Value == "" {
			return
		}
		name := eng.Prefix + strings.ToUpper(nk)
		if out[name] != "" { // an explicit env var wins
			return
		}
		out[name] = e.Value
		exported = append(exported, name)
	})
	if out[eng.Prefix+"ROLES"] == "" {
		out[eng.Prefix+"ROLES"] = filepath.Join(teacherSkillDir, "roles")
	}
	if eng.SkillDir != "" && out["HERDR_SOHO_SKILL_DIR"] == "" {
		out["HERDR_SOHO_SKILL_DIR"] = eng.SkillDir
	}
	out["HERDR_TEACHER_SKILL_DIR"] = teacherSkillDir
	return out, exported
}

// FindSkillDir locates this skill: HERDR_TEACHER_SKILL_DIR, the tree that
// holds the executable (<skill>/bin/herdr-teacher), then
// ~/.agents/skills/herdr-teacher and ~/.claude/skills/herdr-teacher.
func FindSkillDir(env Env, executable string) (string, error) {
	if d := env.Get("HERDR_TEACHER_SKILL_DIR"); d != "" {
		if isSkillDir(d) {
			return d, nil
		}
		return "", errors.New("HERDR_TEACHER_SKILL_DIR has no roles/ and config.defaults: " + d)
	}
	if executable != "" {
		if real, err := filepath.EvalSymlinks(executable); err == nil {
			executable = real
		}
		for d := filepath.Dir(executable); ; d = filepath.Dir(d) {
			if isSkillDir(d) {
				return d, nil
			}
			if filepath.Dir(d) == d {
				break
			}
		}
	}
	home := homeDir(env, runtime.GOOS)
	for _, d := range []string{
		filepath.Join(home, ".agents", "skills", "herdr-teacher"),
		filepath.Join(home, ".claude", "skills", "herdr-teacher"),
	} {
		if isSkillDir(d) {
			if real, err := filepath.EvalSymlinks(d); err == nil {
				return real, nil
			}
			return d, nil
		}
	}
	return "", errors.New("cannot find the herdr-teacher skill directory; set HERDR_TEACHER_SKILL_DIR")
}

func isSkillDir(d string) bool {
	return isFile(filepath.Join(d, "config.defaults")) && isFile(filepath.Join(d, "roles", "reviewer.md"))
}

func isFile(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.Mode().IsRegular()
}
