package PipBot

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/devices/fraction_collector/pipbot"
	"testing"
)

func TestRedisClient_AllKeys(t *testing.T) {
	c := NewRedisClient(":")
	ctx := context.Background()
	nKeys := 1000
	for i := 0; i < nKeys; i++ {
		c.Set(ctx, fmt.Sprintf("test:%d", i), i, 0)
	}
	keyChan := c.AllKeys(ctx, "test")
	keys := make([]string, 0)
	for key := range keyChan {
		keys = append(keys, key)
	}
	if len(keys) != nKeys {
		t.Errorf("expected %d keys, got %d", nKeys, len(keys))
	}
	for i := 0; i < nKeys; i++ {
		c.Del(ctx, fmt.Sprintf("test:%d", i))
	}
}

func TestRedisClient_E2E(t *testing.T) {
	m := pipbot.FluidLevelMap{
		"A1": 0.123,
		"A2": 0.456,
		"A3": 0.789,
	}
	c := NewRedisClient(":")
	ctx := context.Background()
	err := c.Flush(ctx, "test", m)
	if err != nil {
		t.Error(err)
	}
	loaded, err := c.Load(ctx, "test")
	if err != nil {
		t.Error(err)
	}
	if len(loaded) != len(m) {
		t.Errorf("expected %d keys, got %d", len(m), len(loaded))
	}
	for k, v := range m {
		if loaded[k] != v {
			t.Errorf("key %s: expected %f, got %f", k, v, loaded[k])
		}
	}
	for k := range c.AllKeys(ctx, "test") {
		c.Del(ctx, k)
	}
}
