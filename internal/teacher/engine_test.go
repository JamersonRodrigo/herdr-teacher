package teacher

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func lookOnly(found map[string]string) LookPath {
	return func(file string) (string, error) {
		if p, ok := found[file]; ok {
			return p, nil
		}
		return "", errors.New("not found")
	}
}

func touch(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindEngineOrder(t *testing.T) {
	home := t.TempDir()
	teacher := filepath.Join(home, "skills", "herdr-teacher")
	mjs := filepath.Join(home, "skills", "herdr-soho", "scripts", "herdr-soho.mjs")
	touch(t, mjs)
	bin := filepath.Join(home, "soho-bin")
	touch(t, bin)
	env := Env{"HOME": home, "USERPROFILE": home}

	e, err := FindEngine(Env{"HOME": home, "USERPROFILE": home, "HERDR_TEACHER_SOHO_BIN": bin}, teacher, lookOnly(map[string]string{"soho": "/p/soho", "node": "/p/node"}))
	if err != nil || e.Program != bin || e.Prefix != "SOHO_" {
		t.Errorf("explicit bin: %+v %v", e, err)
	}
	e, err = FindEngine(env, teacher, lookOnly(map[string]string{"soho": "/p/soho", "node": "/p/node"}))
	if err != nil || e.Program != "/p/soho" || e.Prefix != "SOHO_" {
		t.Errorf("soho on PATH: %+v %v", e, err)
	}
	e, err = FindEngine(env, teacher, lookOnly(map[string]string{"node": "/p/node"}))
	if err != nil || e.Name != "herdr-soho.mjs" || e.Prefix != "HERDR_SOHO_" || e.Program != "/p/node" {
		t.Fatalf("sibling herdr-soho: %+v %v", e, err)
	}
	if prog, args := e.Command([]string{"roles"}); prog != "/p/node" || len(args) != 2 || !strings.HasSuffix(args[0], "herdr-soho.mjs") || args[1] != "roles" {
		t.Errorf("command: %s %v", prog, args)
	}
	if _, err := FindEngine(env, teacher, lookOnly(nil)); err == nil || !strings.Contains(err.Error(), "neither node nor bun") {
		t.Errorf("no runtime: %v", err)
	}
	empty := t.TempDir()
	if _, err := FindEngine(Env{"HOME": empty, "USERPROFILE": empty}, filepath.Join(empty, "t"), lookOnly(nil)); !errors.Is(err, ErrNoEngine) {
		t.Errorf("nothing: %v", err)
	}
}

func TestEngineEnvPrefixesAndPrecedence(t *testing.T) {
	cfg := &Config{entries: map[string]Entry{}}
	cfg.set("lanes", Entry{Key: "lanes", Value: "on"})
	cfg.set("family_check", Entry{Key: "family_check", Value: "strict"})
	cfg.set("state_dir", Entry{Key: "state_dir", Value: ""})
	cfg.set("teacher_material_dir", Entry{Key: "teacher.material_dir", Value: "aulas"})

	out, exported := EngineEnv(cfg, Env{"SOHO_FAMILY_CHECK": "warn"}, Engine{Prefix: "SOHO_"}, "/t")
	if out["SOHO_LANES"] != "on" || out["SOHO_FAMILY_CHECK"] != "warn" {
		t.Errorf("values: %v", out)
	}
	if _, ok := out["SOHO_STATE_DIR"]; ok {
		t.Error("an empty value must not be exported")
	}
	if _, ok := out["SOHO_TEACHER_MATERIAL_DIR"]; ok {
		t.Error("teacher.* must not be exported")
	}
	if len(exported) != 1 || exported[0] != "SOHO_LANES" {
		t.Errorf("exported: %v", exported)
	}
	if out["SOHO_ROLES"] != filepath.Join("/t", "roles") || out["HERDR_SOHO_SKILL_DIR"] != "" {
		t.Errorf("go engine env: %v", out)
	}
	for k := range out {
		if strings.HasPrefix(k, "HERDR_SOHO_") {
			t.Errorf("the Go engine must not get legacy %s", k)
		}
	}

	out, _ = EngineEnv(cfg, Env{}, Engine{Prefix: "HERDR_SOHO_", SkillDir: "/s"}, "/t")
	if out["HERDR_SOHO_LANES"] != "on" || out["HERDR_SOHO_ROLES"] != filepath.Join("/t", "roles") || out["HERDR_SOHO_SKILL_DIR"] != "/s" {
		t.Errorf("js engine env: %v", out)
	}
}

func TestFindSkillDirFromExecutable(t *testing.T) {
	skill := repoSkillDir(t)
	got, err := FindSkillDir(Env{}, filepath.Join(skill, "bin", "herdr-teacher"))
	if err != nil || got != skill {
		t.Errorf("got %q %v, want %q", got, err, skill)
	}
	if _, err := FindSkillDir(Env{"HERDR_TEACHER_SKILL_DIR": t.TempDir()}, ""); err == nil {
		t.Error("an invalid HERDR_TEACHER_SKILL_DIR must fail")
	}
}

type fakeRunner struct {
	look    map[string]string
	exe     string
	cwd     string
	program string
	args    []string
	env     Env
	code    int
}

func (f *fakeRunner) Run(program string, args []string, env Env) (int, error) {
	f.program, f.args, f.env = program, args, env
	return f.code, nil
}
func (f *fakeRunner) LookPath(file string) (string, error) { return lookOnly(f.look)(file) }
func (f *fakeRunner) Executable() (string, error)          { return f.exe, nil }
func (f *fakeRunner) Getwd() (string, error)               { return f.cwd, nil }

func TestRunPassesThroughAndGuards(t *testing.T) {
	skill := repoSkillDir(t)
	home := t.TempDir()
	env := Env{"HOME": home, "USERPROFILE": home, "XDG_CONFIG_HOME": filepath.Join(home, "cfg"), "HERDR_TEACHER_SKILL_DIR": skill}
	f := &fakeRunner{look: map[string]string{"soho": "/p/soho"}, cwd: t.TempDir(), code: 9}
	var out, errb bytes.Buffer

	if code := Run([]string{"wait", "pesquisa"}, env, f, &out, &errb); code != 9 {
		t.Fatalf("exit code not passed through: %d (%s)", code, errb.String())
	}
	if f.program != "/p/soho" || strings.Join(f.args, " ") != "wait pesquisa" {
		t.Errorf("ran %s %v", f.program, f.args)
	}
	if f.env["SOHO_LANE_REVISAO_ROLES"] != "reviewer" || f.env["SOHO_ROLES"] != filepath.Join(skill, "roles") || f.env["SOHO_ORCHESTRATOR_NAME"] != "professor" {
		t.Errorf("engine env misses the teacher config")
	}

	f.program = ""
	errb.Reset()
	if code := Run([]string{"spawn", "professor"}, env, f, &out, &errb); code != 12 || f.program != "" {
		t.Errorf("spawn professor: code %d, ran %q", code, f.program)
	}
	env2 := Env{}
	for k, v := range env {
		env2[k] = v
	}
	env2["SOHO_LANES"] = "off"
	if code := Run([]string{"dispatch", "criacao", "b.md"}, env2, f, &out, &errb); code != 5 {
		t.Errorf("lanes=off from env: code %d", code)
	}

	envRev := Env{}
	for k, v := range env {
		envRev[k] = v
	}
	envRev["HERDR_TEACHER_REVIEW_LANE"] = "rev"
	f.program = ""
	errb.Reset()
	if code := Run([]string{"dispatch", "rev", "b.md", "--role", "reviewer"}, envRev, f, &out, &errb); code != 9 || f.program != "/p/soho" {
		t.Errorf("dispatch rev with HERDR_TEACHER_REVIEW_LANE=rev: code %d, ran %q", code, f.program)
	}
	f.program = ""
	errb.Reset()
	if code := Run([]string{"dispatch", "revisao", "b.md", "--role", "reviewer"}, envRev, f, &out, &errb); code != 5 || f.program != "" {
		t.Errorf("dispatch revisao with HERDR_TEACHER_REVIEW_LANE=rev: code %d, ran %q", code, f.program)
	}

	out.Reset()
	if code := Run([]string{"teacher-env"}, env, &fakeRunner{cwd: t.TempDir()}, &out, &errb); code != 0 {
		t.Fatalf("teacher-env without engine: %d", code)
	}
	if !strings.Contains(out.String(), "exported to the engine: ") || !strings.Contains(out.String(), "SOHO_LANE_CRIACAO_ROLES") {
		t.Errorf("teacher-env output:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "roles:   "+filepath.Join(skill, "roles")) {
		t.Errorf("teacher-env output lacks roles line:\n%s", out.String())
	}
}
