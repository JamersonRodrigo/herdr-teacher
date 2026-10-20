package teacher

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// repoSkillDir is the skill tree this test runs from (internal/teacher → ../..).
func repoSkillDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestNormalizeKeyMatchesSoho(t *testing.T) {
	cases := map[string]string{
		"lane.pesquisa.roles":       "lane_pesquisa_roles",
		"role.content-creator.kind": "role_content_creator_kind",
		" teacher.material_dir ":    "teacher_material_dir",
		"a.b c!":                    "a_bc",
	}
	for in, want := range cases {
		if got := NormalizeKey(in); got != want {
			t.Errorf("NormalizeKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseConfig(t *testing.T) {
	got := ParseConfig("# c\n\na=1\r\n=x\nb = two words \nnoeq\n")
	want := []KV{{"a", "1"}, {"b", "two words"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestUserConfigPath(t *testing.T) {
	if got := UserConfigPath(Env{"XDG_CONFIG_HOME": "/x"}, "linux"); got != filepath.Join("/x", "herdr-teacher", "config") {
		t.Errorf("xdg: %s", got)
	}
	if got := UserConfigPath(Env{"APPDATA": `C:\A`}, "windows"); got != filepath.Join(`C:\A`, "herdr-teacher", "config") {
		t.Errorf("windows: %s", got)
	}
	if got := UserConfigPath(Env{"HOME": "/h"}, "darwin"); got != filepath.Join("/h", ".config", "herdr-teacher", "config") {
		t.Errorf("posix: %s", got)
	}
}

func TestProjectConfigWinsOverDefaults(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ProjectConfigPath(root), []byte("lane.revisao.kind=codex\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := LoadConfig(repoSkillDir(t), Env{"XDG_CONFIG_HOME": filepath.Join(root, "nouser")}, root)
	e, _ := cfg.Get("lane_revisao_kind")
	if e.Value != "codex" || e.Source != "project" {
		t.Errorf("lane_revisao_kind = %+v", e)
	}
	if e, _ := cfg.Get("lane_criacao_kind"); e.Source != "defaults" {
		t.Errorf("lane_criacao_kind source = %s", e.Source)
	}
}

func TestShippedDefaultsSegregateTheReviewer(t *testing.T) {
	root := t.TempDir()
	cfg := LoadConfig(repoSkillDir(t), Env{"XDG_CONFIG_HOME": filepath.Join(root, "nouser")}, root)
	if cfg.Value("lane_criacao_kind", "") == cfg.Value("lane_revisao_kind", "") {
		t.Error("creator and reviewer run the same CLI by default")
	}
	if got := cfg.Value("lane_revisao_roles", ""); got != "reviewer" {
		t.Errorf("revisao lane roles = %q, want only reviewer", got)
	}
	for _, lane := range []string{"lane_pesquisa_roles", "lane_criacao_roles"} {
		for _, r := range strings.Split(cfg.Value(lane, ""), ",") {
			if r == "reviewer" {
				t.Errorf("%s includes the reviewer", lane)
			}
		}
	}
	if cfg.Value("lanes", "") != "on" || cfg.Value("family_check", "") != "strict" {
		t.Error("defaults must keep lanes=on and family_check=strict")
	}
}

func TestRoleFilesAreValid(t *testing.T) {
	want := map[string]string{
		"professor": "read-only", "researcher": "read-only", "content-creator": "edit",
		"reviewer": "read-only", "debater": "read-only", "evaluator": "read-only", "scenario-analyst": "read-only",
	}
	for role, mode := range want {
		data, err := os.ReadFile(filepath.Join(repoSkillDir(t), "roles", role+".md"))
		if err != nil {
			t.Fatal(err)
		}
		text := strings.ReplaceAll(string(data), "\r\n", "\n")
		if !strings.HasPrefix(text, "---\nname: "+role+"\n") {
			t.Errorf("%s: frontmatter must start with name: %s", role, role)
		}
		if !strings.Contains(text, "\nmode: "+mode+"\n") {
			t.Errorf("%s: want mode: %s", role, mode)
		}
	}
	data, _ := os.ReadFile(filepath.Join(repoSkillDir(t), "roles", "reviewer.md"))
	for _, s := range []string{
		"findings: N (P0 a, P1 b, P2 c, P3 d) | verdict: pass|fail",
		"Auditoria adversária concluída. Tentei ativamente encontrar erros conceituais, técnicos e vieses no material, mas FALHEI em encontrar falhas ou contrafatos. Conteúdo aprovado (PASS).",
	} {
		if !strings.Contains(string(data), s) {
			t.Errorf("reviewer.md lacks %q", s)
		}
	}
}
