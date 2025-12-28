package typutil_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/KarpelesLab/typutil"
)

func TestOffsetGet(t *testing.T) {
	ctx := context.Background()

	t.Run("map[string]any", func(t *testing.T) {
		m := map[string]any{"foo": "bar", "num": 42}

		val, err := typutil.OffsetGet(ctx, m, "foo")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "bar" {
			t.Errorf("expected 'bar', got %v", val)
		}

		val, err = typutil.OffsetGet(ctx, m, "num")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}

		val, err = typutil.OffsetGet(ctx, m, "missing")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil for missing key, got %v", val)
		}
	})

	t.Run("map[string]string", func(t *testing.T) {
		m := map[string]string{"foo": "bar", "hello": "world"}

		val, err := typutil.OffsetGet(ctx, m, "foo")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "bar" {
			t.Errorf("expected 'bar', got %v", val)
		}

		val, err = typutil.OffsetGet(ctx, m, "missing")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "" {
			t.Errorf("expected empty string for missing key, got %v", val)
		}
	})

	t.Run("url.Values", func(t *testing.T) {
		v := url.Values{}
		v.Set("name", "alice")
		v.Add("tags", "a")
		v.Add("tags", "b")

		val, err := typutil.OffsetGet(ctx, v, "name")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "alice" {
			t.Errorf("expected 'alice', got %v", val)
		}

		// Should return first value for multi-value key
		val, err = typutil.OffsetGet(ctx, v, "tags")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "a" {
			t.Errorf("expected 'a', got %v", val)
		}

		// Missing key should return nil
		val, err = typutil.OffsetGet(ctx, v, "missing")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil for missing key, got %v", val)
		}
	})

	t.Run("[]any slice", func(t *testing.T) {
		s := []any{"first", "second", "third"}

		val, err := typutil.OffsetGet(ctx, s, "0")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "first" {
			t.Errorf("expected 'first', got %v", val)
		}

		val, err = typutil.OffsetGet(ctx, s, "2")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "third" {
			t.Errorf("expected 'third', got %v", val)
		}

		// Out of bounds should return nil
		val, err = typutil.OffsetGet(ctx, s, "10")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil for out of bounds, got %v", val)
		}

		// Invalid index should return error
		_, err = typutil.OffsetGet(ctx, s, "not-a-number")
		if err == nil {
			t.Errorf("expected error for invalid index")
		}
	})

	t.Run("generic map with string key via reflection", func(t *testing.T) {
		type customMap map[string]int
		m := customMap{"one": 1, "two": 2}

		val, err := typutil.OffsetGet(ctx, m, "one")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != 1 {
			t.Errorf("expected 1, got %v", val)
		}

		val, err = typutil.OffsetGet(ctx, m, "missing")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil for missing key, got %v", val)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := typutil.OffsetGet(ctx, 42, "key")
		if err == nil {
			t.Errorf("expected error for unsupported type")
		}

		_, err = typutil.OffsetGet(ctx, "string", "key")
		if err == nil {
			t.Errorf("expected error for unsupported type")
		}
	})
}

// offsetGetterImpl implements the offsetGetter interface for testing
type offsetGetterImpl struct {
	data map[string]any
}

func (o *offsetGetterImpl) OffsetGet(ctx context.Context, key string) (any, error) {
	return o.data[key], nil
}

func TestOffsetGetWithInterface(t *testing.T) {
	ctx := context.Background()

	getter := &offsetGetterImpl{
		data: map[string]any{"key1": "value1", "key2": 123},
	}

	val, err := typutil.OffsetGet(ctx, getter, "key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}

	val, err = typutil.OffsetGet(ctx, getter, "key2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != 123 {
		t.Errorf("expected 123, got %v", val)
	}
}

// valueReaderImpl implements the valueReader interface for testing
type valueReaderImpl struct {
	value any
}

func (v *valueReaderImpl) ReadValue(ctx context.Context) (any, error) {
	return v.value, nil
}

func TestOffsetGetWithValueReader(t *testing.T) {
	ctx := context.Background()

	reader := &valueReaderImpl{
		value: map[string]any{"nested": "value"},
	}

	val, err := typutil.OffsetGet(ctx, reader, "nested")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("expected 'value', got %v", val)
	}
}
