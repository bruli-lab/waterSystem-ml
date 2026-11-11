package python

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const stderrTailSize = 2000

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func runCommand(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	outStr := strings.TrimSpace(stdout.String())

	combined := outStr
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return combined, fmt.Errorf("timeout excedit executant %s %v", name, args)
	}

	if err != nil {
		stderrTail := tail(strings.TrimSpace(stderr.String()), stderrTailSize)
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return combined, fmt.Errorf("exit code %d: %s", status.ExitStatus(), stderrTail)
			}
		}
		return combined, fmt.Errorf("%w: %s", err, stderrTail)
	}

	return combined, nil
}
