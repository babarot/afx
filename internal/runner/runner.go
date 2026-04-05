package runner

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"golang.org/x/sync/errgroup"
)

// Package is the minimal interface needed by the runner.
type Package interface {
	GetName() string
}

// TaskFunc is the function executed for each package in parallel.
// It receives the context and the completion channel for progress reporting.
type TaskFunc func(ctx context.Context, completion chan<- Status) error

// Result holds the outcome of a single package execution.
type Result struct {
	Name  string
	Error error
}

// Execute runs taskFn for each package in parallel with progress reporting.
// It handles signal interruption, concurrency limiting, and error aggregation.
func Execute(pkgs []Package, taskFn func(pkg Package) TaskFunc) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.GetName()
	}
	progress := NewProgress(names)
	completion := make(chan Status)
	limit := make(chan struct{}, 16)
	results := make(chan Result)

	go func() {
		progress.Print(completion)
	}()

	eg := errgroup.Group{}
	for _, pkg := range pkgs {
		fn := taskFn(pkg)
		eg.Go(func() error {
			limit <- struct{}{}
			defer func() { <-limit }()
			err := fn(ctx, completion)
			select {
			case results <- Result{Name: pkg.GetName(), Error: err}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}

	go func() {
		_ = eg.Wait()
		close(results)
	}()

	var exit []error
	for result := range results {
		if result.Error != nil {
			exit = append(exit, result.Error)
		}
	}
	if err := eg.Wait(); err != nil {
		log.Printf("[ERROR] execution failed: %s", err)
		exit = append(exit, err)
	}

	return errors.Join(exit...)
}
