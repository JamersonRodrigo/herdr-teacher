package teacher

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
)

// Runner starts the engine and returns its exit code.
type Runner interface {
	Run(program string, args []string, env Env) (int, error)
	LookPath(file string) (string, error)
	Executable() (string, error)
	Getwd() (string, error)
}

// OSRunner runs the engine as a child process with the terminal inherited.
type OSRunner struct{}

func (OSRunner) Run(program string, args []string, env Env) (int, error) {
	cmd := exec.Command(program, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = env.List()
	// Ctrl-C reaches the engine through the terminal; this process stays
	// alive until the engine exits so the exit code is passed through.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	defer signal.Stop(sig)
	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitCode(exitErr), nil
	}
	return 0, err
}

func (OSRunner) LookPath(file string) (string, error) { return exec.LookPath(file) }
func (OSRunner) Executable() (string, error)          { return os.Executable() }
func (OSRunner) Getwd() (string, error)               { return os.Getwd() }

// Run is the whole command: teacher-env, or the guards and then the engine.
func Run(argv []string, env Env, r Runner, stdout, stderr io.Writer) int {
	fail := func(code int, msg string) int {
		fmt.Fprintf(stderr, "herdr-teacher: %s\n", msg)
		return code
	}
	exe, _ := r.Executable()
	skillDir, err := FindSkillDir(env, exe)
	if err != nil {
		return fail(2, err.Error())
	}
	cwd, err := r.Getwd()
	if err != nil {
		return fail(2, err.Error())
	}
	cfg := LoadConfig(skillDir, env, cwd)
	eng, engErr := FindEngine(env, skillDir, r.LookPath)

	if len(argv) > 0 && argv[0] == "teacher-env" {
		printTeacherEnv(stdout, cfg, env, eng, engErr, skillDir, cwd)
		return 0
	}
	if engErr != nil {
		return fail(2, engErr.Error())
	}
	engineEnv, _ := EngineEnv(cfg, env, eng, skillDir)
	lanes := engineEnv.Get(eng.Prefix + "LANES")
	if lanes == "" {
		lanes = "on"
	}
	reviewLane := env.Get("HERDR_TEACHER_REVIEW_LANE")
	if reviewLane == "" {
		reviewLane = cfg.Value("teacher_review_lane", "revisao")
	}
	if ref := Guard(argv, lanes, reviewLane); ref != nil {
		return fail(ref.Code, ref.Msg)
	}
	program, args := eng.Command(argv)
	code, err := r.Run(program, args, engineEnv)
	if err != nil {
		return fail(2, "could not run the engine: "+err.Error())
	}
	return code
}

func printTeacherEnv(w io.Writer, cfg *Config, env Env, eng Engine, engErr error, skillDir, cwd string) {
	engine := eng.Name + " (" + eng.Program + ")"
	if engErr != nil {
		engine = "(not found: " + engErr.Error() + ")"
		eng.Prefix = "SOHO_"
	}
	fmt.Fprintf(w, "skill:   %s\n", skillDir)
	fmt.Fprintf(w, "engine:  %s\n", engine)
	fmt.Fprintf(w, "prefix:  %s\n", eng.Prefix)
	fmt.Fprintf(w, "roles:   %s\n", filepath.Join(skillDir, "roles"))
	fmt.Fprintf(w, "user:    %s\n", UserConfigPath(env, runtime.GOOS))
	fmt.Fprintf(w, "project: %s\n\n", ProjectConfigPath(ProjectRoot(cwd)))
	cfg.Each(func(_ string, e Entry) {
		fmt.Fprintf(w, "%s=%s  # %s\n", e.Key, e.Value, e.Source)
	})
	_, exported := EngineEnv(cfg, env, eng, skillDir)
	list := "(none)"
	if len(exported) > 0 {
		list = strings.Join(exported, " ")
	}
	fmt.Fprintf(w, "\nexported to the engine: %s\n", list)
}
