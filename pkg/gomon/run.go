package gomon

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

type IO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (io *IO) stdin() io.Reader {
	if io == nil || io.Stdin == nil {
		return os.Stdin
	}
	return io.Stdin
}

func (io *IO) stdout() io.Writer {
	if io == nil || io.Stdout == nil {
		return os.Stdout
	}
	return io.Stdout
}

func (io *IO) stderr() io.Writer {
	if io == nil || io.Stderr == nil {
		return os.Stderr
	}
	return io.Stderr
}

func RunJob(ctx context.Context, job Job, io *IO) error {
	parts := job.Exec.Parts
	if len(parts) == 0 {
		return fmt.Errorf("no exec commands")
	}
	if len(parts) == 1 {
		return runShell(ctx, parts[0], io)
	}
	if !job.Parallel {
		for _, p := range parts {
			if err := runShell(ctx, p, io); err != nil {
				return err
			}
		}
		return nil
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(parts))
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, p := range parts {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runShell(ctx2, p, io); err != nil {
				select {
				case errCh <- err:
				default:
				}
				cancel()
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func runShell(ctx context.Context, script string, io *IO) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", script)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", script)
	}
	cmd.Stdout = io.stdout()
	cmd.Stderr = io.stderr()
	cmd.Stdin = io.stdin()
	setProcGroup(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		killProcessTree(cmd)
		return ctx.Err()
	}
	return waitErr
}

func OnInterrupt(cancel context.CancelFunc) (stop func()) {
	ch := make(chan os.Signal, 4)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return func() { signal.Stop(ch) }
}
