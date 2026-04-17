package cache

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c := New[string](time.Minute)
	if c == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestGetMissing(t *testing.T) {
	c := New[string](time.Minute)
	_, exists, fresh := c.Get("missing")
	if exists || fresh {
		t.Errorf("expected exists=false, fresh=false; got exists=%v, fresh=%v", exists, fresh)
	}
}

func TestSetAndGetFresh(t *testing.T) {
	c := New[string](time.Minute)
	c.Set("key", "value")

	val, exists, fresh := c.Get("key")
	if !exists || !fresh {
		t.Fatalf("expected exists=true, fresh=true; got exists=%v, fresh=%v", exists, fresh)
	}
	if val != "value" {
		t.Errorf("expected 'value', got %q", val)
	}
}

func TestGetStale(t *testing.T) {
	c := New[string](1 * time.Millisecond)
	c.Set("key", "value")

	time.Sleep(5 * time.Millisecond)

	val, exists, fresh := c.Get("key")
	if !exists {
		t.Fatal("expected exists=true for stale entry")
	}
	if fresh {
		t.Error("expected fresh=false for expired entry")
	}
	if val != "value" {
		t.Errorf("expected 'value', got %q", val)
	}
}

func TestSetOverwrites(t *testing.T) {
	c := New[string](time.Minute)
	c.Set("key", "first")
	c.Set("key", "second")

	val, _, _ := c.Get("key")
	if val != "second" {
		t.Errorf("expected 'second', got %q", val)
	}
}

func TestMultipleKeys(t *testing.T) {
	c := New[int](time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)

	a, _, _ := c.Get("a")
	b, _, _ := c.Get("b")
	if a != 1 || b != 2 {
		t.Errorf("expected a=1, b=2; got a=%d, b=%d", a, b)
	}
}
