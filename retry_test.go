package retry

import (
	"math/rand"
	"testing"
	"time"
)

func TestRetryMultiple(t *testing.T) {
	rand.Seed(time.Now().Unix())

	var (
		total = rand.Intn(10) + 1
		count int
	)

	New(Sleep(1 * time.Millisecond)).Do(func() (Reason, error) {
		if count == total {
			return Stop, nil
		}

		count++
		return Again, nil
	})

	if count != total {
		t.Fatalf("Retry should've returned %v, got %v instead.", total, count)
	}
}

func TestRetryMaxAttempts(t *testing.T) {
	rand.Seed(time.Now().Unix())

	total := rand.Intn(20) + 1
	count := 0
	r := New(Sleep(1*time.Millisecond), MaxAttempts(total))

	err := r.Do(func() (Reason, error) {
		// Increase the counter in 1 on every iteration
		count++

		// try over and over and over again, it should
		// stop after 5 attempts
		return Again, nil
	})

	if err == nil {
		t.Fatalf("Retry expected to return an error, got nothing instead")
	}

	if err != ErrRetriesExhausted {
		t.Fatalf("Retry expected to return error %q, got %q", ErrRetriesExhausted.Error(), err.Error())
	}

	if r.MaxAttempts() != total {
		t.Fatalf("Retry expected MaxAttempts to be %d, but function returned %d", total, r.MaxAttempts())
	}

	if count != total {
		t.Fatalf("Retry should've counted %v attempts, but did %v instead.", total, count)
	}
}
