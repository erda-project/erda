// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package polling provides utilities for polling operations with exponential backoff.
package polling

import (
	"context"
	"fmt"
	"time"
)

// Config defines the configuration for polling with exponential backoff.
type Config struct {
	// InitialInterval is the initial wait interval before the first retry.
	InitialInterval time.Duration

	// MaxInterval is the maximum wait interval between retries.
	MaxInterval time.Duration

	// Timeout is the maximum total time for the entire polling operation.
	// If zero, polling will continue indefinitely until success or context cancellation.
	Timeout time.Duration

	// Multiplier is the factor by which the interval increases after each retry.
	// Default is 2.0 if not set.
	Multiplier float64
}

// DefaultConfig returns a default polling configuration.
func DefaultConfig() Config {
	return Config{
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		Timeout:         2 * time.Minute,
		Multiplier:      2.0,
	}
}

// Result represents the result of a single poll attempt.
type Result struct {
	// Done indicates whether polling should stop (either success or permanent failure).
	Done bool

	// Data is the result data when polling succeeds.
	Data any

	// Err is set when a permanent error occurs (polling should stop).
	Err error
}

// PollFunc is the function called on each poll attempt.
// It should return a Result indicating whether to continue polling.
type PollFunc func(ctx context.Context) Result

// Poll executes the given function with exponential backoff until:
// - The function returns Done=true (success or permanent failure)
// - The timeout is reached
// - The context is cancelled
//
// Returns the final result data and error.
func Poll(ctx context.Context, cfg Config, fn PollFunc) (any, error) {
	if cfg.Multiplier == 0 {
		cfg.Multiplier = 2.0
	}

	// Create a timeout context if timeout is specified
	var cancel context.CancelFunc
	if cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	interval := cfg.InitialInterval
	for {
		// Check if context is cancelled or timed out before polling
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("polling timeout after %v", cfg.Timeout)
			}
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		// Execute the poll function
		result := fn(ctx)
		if result.Done {
			return result.Data, result.Err
		}

		// Wait with exponential backoff, but respect context cancellation
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("polling timeout after %v", cfg.Timeout)
			}
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-time.After(interval):
		}

		// Exponential backoff: multiply the interval, but cap at MaxInterval
		interval = time.Duration(float64(interval) * cfg.Multiplier)
		if interval > cfg.MaxInterval {
			interval = cfg.MaxInterval
		}
	}
}
