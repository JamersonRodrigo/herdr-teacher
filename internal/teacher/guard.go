package teacher

import (
	"fmt"
	"strings"
)

// Roles only the professor (the pane running the skill) plays.
var orchestratorRoles = map[string]bool{"professor": true, "planner": true}

// Roles whose output the reviewer audits; the review pane never plays them.
var authorRoles = map[string]bool{"content-creator": true, "researcher": true}

// Flags that consume the next argument as their value when looking for the
// first positional argument (role for spawn/run, agent for dispatch).
var valueFlags = map[string]bool{
	"--role":      true,
	"--timeout":   true,
	"--for":       true,
	"--name":      true,
	"--kind":      true,
	"--model":     true,
	"--effort":    true,
	"--approvals": true,
	"--cwd":       true,
	"--pane":      true,
	"--direction": true,
	"--ratio":     true,
	"--tab-label": true,
	"--mcp":       true,
}

const reviewRole = "reviewer"

// Refusal is a guard decision: the message and the exit code.
type Refusal struct {
	Code int
	Msg  string
}

// Guard returns a refusal for a command that would break the segregation
// rules or rewrite the engine's own files, or nil. lanes is the effective
// `lanes` value and reviewLane the name of the reviewer's lane.
func Guard(argv []string, lanes, reviewLane string) *Refusal {
	if len(argv) == 0 {
		return nil
	}
	cmd, rest := argv[0], argv[1:]
	if lanes == "off" && (cmd == "spawn" || cmd == "run" || cmd == "dispatch") {
		return &Refusal{5, "lanes=off would let the reviewer reuse a researcher or content-creator pane. herdr-teacher requires lanes=on."}
	}
	// These write the engine's own files with the teacher lanes in effect:
	// `doctor --fix` would delete the user's soho lanes (build/review do not
	// exist here), `setup` writes the soho block and hooks, and `config set`
	// writes the soho config. The teacher config is edited by hand.
	if cmd == "doctor" && contains(rest, "--fix") {
		return &Refusal{2, "doctor --fix would rewrite your soho config against the teacher lanes. Run plain `doctor`; fix soho with its own command."}
	}
	if cmd == "setup" {
		return &Refusal{2, "setup writes soho files. Edit the teacher config instead (see: herdr-teacher teacher-env)."}
	}
	if cmd == "config" && len(rest) > 0 && rest[0] == "set" {
		return &Refusal{2, "config set writes soho files. Edit the teacher config instead (see: herdr-teacher teacher-env)."}
	}
	if cmd == "spawn" || cmd == "run" {
		role := firstPositional(rest)
		if orchestratorRoles[role] {
			return &Refusal{12, fmt.Sprintf("spawn %s: the professor is the agent running this skill and does not open a pane. Teach in this session.", role)}
		}
		if role == reviewRole && contains(rest, "--name") {
			return &Refusal{5, fmt.Sprintf("spawn reviewer --name: keep the lane name (%s) so the segregation guard can recognize the reviewer pane.", reviewLane)}
		}
	}
	if cmd == "dispatch" {
		agent := firstPositional(rest)
		role := flagValue(rest, "--role")
		isReviewPane := agent == reviewLane || strings.HasPrefix(agent, reviewLane+"-")
		if role == reviewRole && !isReviewPane {
			return &Refusal{5, fmt.Sprintf("dispatch %s --role reviewer: the reviewer runs only in its own pane ('%s'), never where research or content was produced.", agent, reviewLane)}
		}
		if authorRoles[role] && isReviewPane {
			return &Refusal{5, fmt.Sprintf("dispatch %s --role %s: the review pane never produces the material it audits.", agent, role)}
		}
	}
	return nil
}

// The role (spawn/run) and the agent (dispatch) are the first positional
// argument. Flags taking values consume the next token; `--` alone stops
// the search.
func firstPositional(args []string) string {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		if valueFlags[a] {
			i++
			continue
		}
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}

func flagValue(args []string, flag string) string {
	last := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == flag && i+1 < len(args) {
			last = args[i+1]
			i++
			continue
		}
		if v, ok := strings.CutPrefix(a, flag+"="); ok {
			last = v
		}
	}
	return last
}

func contains(args []string, s string) bool {
	for _, a := range args {
		if a == s {
			return true
		}
	}
	return false
}
