// Command herdr-teacher runs the soho engine with the herdr-teacher roles,
// lanes and segregation guards.
package main

import (
	"os"

	"github.com/JamersonRodrigo/herdr-teacher/internal/teacher"
)

func main() {
	os.Exit(teacher.Run(os.Args[1:], teacher.EnvFromOS(), teacher.OSRunner{}, os.Stdout, os.Stderr))
}
