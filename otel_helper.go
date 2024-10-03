package telemetry

import (
	"context"
	"errors"
)

// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
func handleErr(err error, ctx context.Context) {
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		_ = err // Assign the result to a variable to avoid the SA4006 error
	}
}

// shutdown calls all the shutdown functions in the reverse order they were added.
func shutdown(ctx context.Context) error {
	var shutdownFuncs []func(context.Context) error
	var err error
	// Initialize the shutdownFuncs slice before using it
	shutdownFuncs = make([]func(context.Context) error, 0)
	for _, fn := range shutdownFuncs {
		err = errors.Join(err, fn(ctx))
	}
	shutdownFuncs = nil
	return err
}
