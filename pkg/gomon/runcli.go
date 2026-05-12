package gomon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

// RunCLI runs the gomon command-line. Pass os.Args[1:] for args (job name; flags reserved).
// It uses the current working directory from getwd (pass [os.Getwd] from main).
func RunCLI(args []string, getwd func() (string, error), io *IO) error {
	if io == nil {
		io = &IO{}
	}
	if len(args) == 0 {
		fmt.Fprintf(io.stderr(), "usage: gomon <job>\n")
		return errors.New("missing job name")
	}
	if args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintf(io.stdout(), "usage: gomon <job>\n")
		return nil
	}
	jobName := args[0]
	root, err := getwd()
	if err != nil {
		return err
	}
	jobs, cfgPath, err := LoadConfig(root)
	if err != nil {
		return err
	}
	job, ok := jobs[jobName]
	if !ok {
		return fmt.Errorf("unknown job %q (config %s)", jobName, filepath.Base(cfgPath))
	}
	ApplyDefaults(&job)

	if len(job.Watch) == 0 {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		stop := OnInterrupt(cancel)
		defer stop()
		return RunJob(ctx, job, io)
	}
	return RunWatch(context.Background(), root, job, io)
}

// RunWatch polls the filesystem and restarts the job when watched files change.
// It honors SIGINT/SIGTERM by cancelling the root context.
func RunWatch(parent context.Context, root string, job Job, io *IO) error {
	if io == nil {
		io = &IO{}
	}
	rootCtx, rootCancel := context.WithCancel(parent)
	defer rootCancel()
	stop := OnInterrupt(rootCancel)
	defer stop()

	prev, err := SnapshotMTimes(root, job)
	if err != nil {
		return err
	}

	for {
		jobCtx, jobCancel := context.WithCancel(rootCtx)
		done := make(chan error, 1)
		go func() {
			done <- RunJob(jobCtx, job, io)
		}()

		ticker := time.NewTicker(300 * time.Millisecond)
	watchLoop:
		for {
			select {
			case err := <-done:
				ticker.Stop()
				jobCancel()
				if rootCtx.Err() != nil {
					return nil
				}
				if err != nil && !errors.Is(err, context.Canceled) {
					fmt.Fprintf(io.stderr(), "gomon: command exited: %v\n", err)
				}
				if !errors.Is(err, context.Canceled) {
					time.Sleep(200 * time.Millisecond)
				}
				break watchLoop
			case <-ticker.C:
				next, err := SnapshotMTimes(root, job)
				if err != nil {
					continue
				}
				if !mapsEqualInt64(prev, next) {
					ticker.Stop()
					prev = next
					jobCancel()
					<-done
					fmt.Fprintf(io.stderr(), "gomon: restarting...\n")
					break watchLoop
				}
			case <-rootCtx.Done():
				ticker.Stop()
				jobCancel()
				<-done
				return nil
			}
		}
	}
}

// DiscardIO returns an [IO] that sends stdout/stderr to [io.Discard] and stdin to an empty reader.
func DiscardIO() *IO {
	return &IO{
		Stdin:  strings.NewReader(""),
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
}
