package retry

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ErrOSRequestedCancellation gets returned from the Do() operation
// when the OS sends a termination signal
var ErrOSRequestedCancellation = errors.New("received OS signal to stop operation")

// ErrRetriesExhausted gets returned from the Do() operation when
// the maximum amount of retries gets exhausted and no more retries
// are attempted
var ErrRetriesExhausted = errors.New("exhausted the number of retries")

// Reason is a custom type to pass to
// retry in order to either try again or stop trying
type Reason int

const (
	// Stop will instruct the retry function to stop trying
	Stop Reason = iota

	// Again will instruct the retry function to execute the
	// action again, sleeping for a certain amount of time before
	// executing it again
	Again
)

// Operation is a function that can execute a code like
// Javascript callbacks, and return a reason to repeat the same
// action again or just stop and continue.
type Operation func() (Reason, error)

// Config is a custom type for functions that accept retry
type Config func(*Retry)

// Retry is a package that allows to execute a callback function multiple
// times until that function thinks it's safe to stop. The flow is handled by returning
// a reason, either Retry or Stop.
type Retry struct {
	sleep       time.Duration
	maxattempts int
}

// MaxAttempts return the currently set number of attempts
func (r *Retry) MaxAttempts() int {
	return r.maxattempts
}

// New creates a new retry operation with a default
// timeout of 5 seconds per iteration, which you can
// change by passing a RetryConfig with the Sleep option.
func New(options ...Config) *Retry {
	current := Retry{
		sleep: 5 * time.Second,
	}

	for _, opt := range options {
		opt(&current)
	}

	return &current
}

// Sleep is a retry configuration you can pass to the New()
// function to change the default sleep time per iteration.
func Sleep(d time.Duration) Config {
	return func(r *Retry) {
		r.sleep = d
	}
}

// MaxAttempts is a retry configuration you can pass to the New()
// function to change the default amount of retries the package
// can do. By default is unlimited.
func MaxAttempts(attempts int) Config {
	return func(r *Retry) {
		r.maxattempts = attempts
	}
}

// Do retries a RetryOperation until either retry.Stop or retry.Again
// gets returned. Stop will continue the parent flow and Again will execute
// the RetryOperation again.
func (r *Retry) Do(fn Operation) error {
	// Wait until you get something here
	close := make(chan error, 1)

	// Create a goroutine to handle the operation loop
	go func(callback Operation, ch chan error) {
		// Start with a single attempt
		attempts := 1
		for {
			// If the attempts were set to something other than zero
			// and we hit the maximum number of attempts, then exhaust
			// and return
			if r.maxattempts != 0 && attempts > r.maxattempts {
				ch <- ErrRetriesExhausted
				return
			}

			// Call the user-given function
			reason, err := callback()

			// Check the returned values
			switch reason {
			case Stop:
				ch <- err
				return
			case Again:
				time.Sleep(r.sleep)
				attempts++
				continue
			}
		}
	}(fn, close)

	// This goroutine will wait for a OS Signal to come by
	go func(ch chan error) {
		// Create an OS signal, so if terraform tells us to stop
		// we don't try again. This will wait until fn() goes to the next loop
		finish := make(chan os.Signal, 1)
		signal.Notify(finish, os.Interrupt, syscall.SIGTERM)

		// This will block here until we get a signal, it won't
		// go to the line below unless the signal happen
		<-finish

		// If we reach this line, it means we did received a signal, so we need
		// to exit per OS request
		ch <- ErrOSRequestedCancellation
	}(close)

	// Retrieve the value we will return
	return <-close
}
