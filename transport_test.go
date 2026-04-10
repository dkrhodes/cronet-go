package cronet

import (
	"context"
	"testing"
	"time"
)

func TestURLResponseFinishIsIdempotent(t *testing.T) {
	r := &urlResponse{
		read:   make(chan int),
		cancel: make(chan struct{}),
		done:   make(chan struct{}),
	}
	r.wg.Add(1)

	r.finish(URLRequest{}, nil)
	r.close(URLRequest{}, context.Canceled)

	waitDone := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("waitgroup did not complete")
	}

	select {
	case <-r.done:
	default:
		t.Fatal("done channel was not closed")
	}

	if r.err != nil {
		t.Fatalf("expected nil error after successful finish, got %v", r.err)
	}
}
