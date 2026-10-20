package teacher

import "testing"

func TestGuard(t *testing.T) {
	cases := []struct {
		argv  []string
		lanes string
		code  int // 0 = allowed
	}{
		{[]string{"spawn", "professor"}, "on", 12},
		{[]string{"run", "planner", "b.md"}, "on", 12},
		{[]string{"spawn", "reviewer", "--name", "x"}, "on", 5},
		{[]string{"spawn", "reviewer"}, "on", 0},
		{[]string{"dispatch", "pesquisa", "b.md", "--role", "reviewer"}, "on", 5},
		{[]string{"dispatch", "criacao", "b.md", "--role=reviewer"}, "on", 5},
		{[]string{"dispatch", "revisao", "b.md", "--role", "content-creator"}, "on", 5},
		{[]string{"dispatch", "revisao-2", "b.md", "--role", "researcher"}, "on", 5},
		{[]string{"dispatch", "revisao", "b.md", "--role", "reviewer"}, "on", 0},
		{[]string{"dispatch", "revisaoX", "b.md", "--role", "reviewer"}, "on", 5},
		{[]string{"dispatch", "pesquisa", "b.md", "--role", "debater"}, "on", 0},
		{[]string{"dispatch", "--role", "content-creator", "revisao", "b.md"}, "on", 5},
		{[]string{"dispatch", "--timeout", "900000", "pesquisa", "b.md", "--role", "reviewer"}, "on", 5},
		{[]string{"dispatch", "--role=reviewer", "revisao", "b.md"}, "on", 0},
		{[]string{"dispatch", "pesquisa", "b.md", "--role", "content-creator", "--role", "reviewer"}, "on", 5},
		{[]string{"dispatch", "revisao", "b.md", "--role", "reviewer", "--role", "researcher"}, "on", 5},
		{[]string{"dispatch", "pesquisa", "b.md", "--role", "reviewer", "--role", "debater"}, "on", 0},
		{[]string{"dispatch", "revisao", "b.md", "--role=researcher", "--role", "reviewer"}, "on", 0},
		{[]string{"spawn", "--kind", "claude", "professor"}, "on", 12},
		{[]string{"spawn", "--fresh", "professor"}, "on", 12},
		{[]string{"spawn", "--", "professor"}, "on", 0},
		{[]string{"spawn", "researcher"}, "off", 5},
		{[]string{"doctor", "--fix"}, "on", 2},
		{[]string{"doctor"}, "on", 0},
		{[]string{"setup", "--dry-run"}, "on", 2},
		{[]string{"config", "set", "a", "b"}, "on", 2},
		{[]string{"config"}, "on", 0},
		{[]string{"roster"}, "off", 0},
		{nil, "on", 0},
	}
	for _, c := range cases {
		ref := Guard(c.argv, c.lanes, "revisao")
		got := 0
		if ref != nil {
			got = ref.Code
		}
		if got != c.code {
			t.Errorf("Guard(%v, lanes=%s) = %d, want %d", c.argv, c.lanes, got, c.code)
		}
	}
}
