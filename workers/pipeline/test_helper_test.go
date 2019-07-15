package pipeline

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

var buildKillContext = false

func fakeExecCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if buildKillContext {
		c, cancel := context.WithTimeout(context.Background(), 0)
		defer cancel()
		ctx = c
	}
	cs := []string{"-test.run=TestExecCommandContextHelper", "--", name}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	arg := strings.Join(cs, ",")
	envArgs := os.Getenv("CMD_ARGS")
	if len(envArgs) != 0 {
		envArgs += ":" + arg
	} else {
		envArgs = arg
	}
	_ = os.Setenv("CMD_ARGS", envArgs)
	return cmd
}

func TestExecCommandContextHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}
