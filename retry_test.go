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
		} else {
			count++
			return Again, nil
		}
	})

	if count != total {
		t.Fatalf("Retry should've returned %v, got %v instead.", total, count)
	}
}
