package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	data := 0
	err, _ := Retry(context.Background(), func() error {
		data++
		return nil
	}, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data != 1 {
		t.Fatalf("expected data to be 1, got %d", data)
	}
}

func TestRetry02(t *testing.T) {
	data := 0
	err, _ := Retry(context.Background(), func() error {
		data++
		if data < 2 {
			return errors.New("retry")
		}
		return nil
	}, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data != 2 {
		t.Fatalf("expected data to be 1, got %d", data)
	}
}

func TestRetry03(t *testing.T) {
	data := 0
	err, _ := Retry(context.Background(), func() error {
		data++
		if data < 4 {
			return errors.New("retry")
		}
		return nil
	}, 3)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if data != 3 {
		t.Fatalf("Data is not matched. expected=%d, actual=%d", 3, data)
	}
}

func TestRetry04(t *testing.T) {
	data := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*18)
	defer cancel()

	err, _ := Retry(ctx, func() error {
		data++
		time.Sleep(time.Millisecond * 10)
		return errors.New("retry")
	}, 3)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected error is context.DeadlineExceeded, actual=%v", err)
	}
	if data != 2 {
		t.Fatalf("Data is not matched. expected=%d, actual=%d", 2, data)
	}
}

func TestRetry05(t *testing.T) {
	data := 0
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond * 15)
		cancel()
	}()

	err, _ := Retry(ctx, func() error {
		data++
		time.Sleep(time.Millisecond * 10)
		return errors.New("retry")
	}, 3)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected error is context.DeadlineExceeded, actual=%v", err)
	}
	if data != 2 {
		t.Fatalf("Data is not matched. expected=%d, actual=%d", 2, data)
	}
}

func TestRetry06(t *testing.T) {
	data := 0

	err, _ := Retry(context.Background(), func() error {
		data++
		return context.Canceled
	}, 3)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected error is context.DeadlineExceeded, actual=%v", err)
	}
	if data != 1 {
		t.Fatalf("Data is not matched. expected=%d, actual=%d", 1, data)
	}
}

func TestRetry07(t *testing.T) {
	data := 0

	err, _ := Retry(context.Background(), func() error {
		data++
		return context.DeadlineExceeded
	}, 3)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected error is context.DeadlineExceeded, actual=%v", err)
	}
	if data != 1 {
		t.Fatalf("Data is not matched. expected=%d, actual=%d", 1, data)
	}
}
