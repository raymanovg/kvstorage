package cache

import "testing"

func TestLRUCacheEviction(t *testing.T) {
	c := NewLRUCache[string, string](2)

	if err := c.Put("a", "1"); err != nil {
		t.Fatalf("put a: %v", err)
	}
	if err := c.Put("b", "2"); err != nil {
		t.Fatalf("put b: %v", err)
	}
	if err := c.Put("c", "3"); err != nil {
		t.Fatalf("put c: %v", err)
	}

	if c.Len() != 2 {
		t.Fatalf("expected len 2 got %d", c.Len())
	}
	if _, err := c.Get("a"); err == nil {
		t.Fatalf("expected key 'a' to be evicted")
	}
	if v, err := c.Get("c"); err != nil || v != "3" {
		t.Fatalf("unexpected get c result %v %v", v, err)
	}
}
