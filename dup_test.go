package typutil_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/KarpelesLab/typutil"
)

type dupTestSruct struct {
	A  []byte
	B  string
	C  *int
	d  *int
	e  *int
	E  map[string]string
	F  []string
	X  any
	Y  any
	z  string
	t1 time.Time
	t2 time.Time
}

func TestDup(t *testing.T) {
	v := 42
	w := 1337

	loc := time.FixedZone("UTC-8", -8*60*60)
	a := &dupTestSruct{
		A:  []byte("hello"),
		B:  "world",
		C:  &v,
		d:  &w,
		e:  &v,
		E:  map[string]string{"foo": "bar"},
		X:  w,
		z:  "are you here?",
		t1: time.Now().In(loc),
		t2: time.Now().In(loc),
	}

	b := typutil.DeepClone(*a)

	if !bytes.Equal(a.A, b.A) {
		t.Errorf("b should be equal a")
	}

	if b.B != "world" {
		t.Errorf("b.B should equal world")
	}

	b.A[0] = 'H'

	if bytes.Equal(a.A, b.A) {
		t.Errorf("b should not be equal a")
	}

	if a.C == b.C {
		t.Errorf("a.C should not equal b.C")
	}
	if b.C != b.e {
		t.Errorf("b.C should equal b.e (same pointer)")
	}

	*b.C = 43

	if v != 42 {
		t.Errorf("b.C should not affect v")
	}

	if b.d == nil {
		t.Errorf("b.d should not be nil")
	} else if a.d == b.d {
		t.Errorf("a.d should not equal b.d")
	} else if *b.d != 1337 {
		t.Errorf("b.d should equal 1337")
	}
}

type cloneSkipStruct struct {
	Name    string
	Value   int
	Skipped *int `clone:"-"`
	Data    []byte
}

func TestDeepCloneSkipTag(t *testing.T) {
	v := 42
	original := &cloneSkipStruct{
		Name:    "test",
		Value:   100,
		Skipped: &v,
		Data:    []byte("hello"),
	}

	cloned := typutil.DeepClone(*original)

	// Regular fields should be cloned
	if cloned.Name != "test" {
		t.Errorf("cloned.Name should equal 'test', got %q", cloned.Name)
	}
	if cloned.Value != 100 {
		t.Errorf("cloned.Value should equal 100, got %d", cloned.Value)
	}
	if !bytes.Equal(cloned.Data, []byte("hello")) {
		t.Errorf("cloned.Data should equal 'hello'")
	}

	// Skipped field should retain shallow copy (same pointer as original)
	if cloned.Skipped != original.Skipped {
		t.Errorf("cloned.Skipped should be same pointer as original.Skipped (clone:\"-\" should skip deep clone)")
	}

	// Verify Data was actually deep cloned (not same slice)
	cloned.Data[0] = 'H'
	if bytes.Equal(original.Data, cloned.Data) {
		t.Errorf("original.Data should not be affected by changes to cloned.Data")
	}
}

func TestDeepCloneInterfacePreservesType(t *testing.T) {
	// Test that cloning an interface preserves the underlying type
	// (not wrapping it in a pointer)

	t.Run("int value in interface", func(t *testing.T) {
		var original any = 42
		cloned := typutil.DeepClone(original)

		// Check that cloned value is still an int, not *int
		if _, ok := cloned.(int); !ok {
			t.Errorf("cloned should be int, got %T", cloned)
		}
		if cloned.(int) != 42 {
			t.Errorf("cloned should equal 42, got %v", cloned)
		}
	})

	t.Run("string value in interface", func(t *testing.T) {
		var original any = "hello"
		cloned := typutil.DeepClone(original)

		if _, ok := cloned.(string); !ok {
			t.Errorf("cloned should be string, got %T", cloned)
		}
		if cloned.(string) != "hello" {
			t.Errorf("cloned should equal 'hello', got %v", cloned)
		}
	})

	t.Run("struct value in interface", func(t *testing.T) {
		type testStruct struct {
			Value int
		}
		var original any = testStruct{Value: 123}
		cloned := typutil.DeepClone(original)

		if _, ok := cloned.(testStruct); !ok {
			t.Errorf("cloned should be testStruct, got %T", cloned)
		}
		if cloned.(testStruct).Value != 123 {
			t.Errorf("cloned.Value should equal 123, got %v", cloned.(testStruct).Value)
		}
	})

	t.Run("pointer in interface", func(t *testing.T) {
		v := 42
		var original any = &v
		cloned := typutil.DeepClone(original)

		// Should still be *int
		clonedPtr, ok := cloned.(*int)
		if !ok {
			t.Errorf("cloned should be *int, got %T", cloned)
		}
		// But should be a different pointer
		if clonedPtr == &v {
			t.Errorf("cloned pointer should not be same as original")
		}
		if *clonedPtr != 42 {
			t.Errorf("cloned value should equal 42, got %d", *clonedPtr)
		}
	})

	t.Run("slice in interface", func(t *testing.T) {
		original := []int{1, 2, 3}
		var iface any = original
		cloned := typutil.DeepClone(iface)

		clonedSlice, ok := cloned.([]int)
		if !ok {
			t.Errorf("cloned should be []int, got %T", cloned)
		}
		if len(clonedSlice) != 3 {
			t.Errorf("cloned slice should have 3 elements")
		}
		// Modify cloned to verify it's a deep copy
		clonedSlice[0] = 999
		if original[0] == 999 {
			t.Errorf("modifying cloned slice should not affect original")
		}
	})
}

// selfRefNode is used to test self-referential pointer structures
type selfRefNode struct {
	Value int
	Self  *selfRefNode
	Next  *selfRefNode
}

func TestDeepCloneSelfReferentialPointer(t *testing.T) {
	// Test that self-referential structures don't cause infinite recursion
	// and are cloned correctly

	t.Run("simple self-reference", func(t *testing.T) {
		node := &selfRefNode{Value: 1}
		node.Self = node // points to itself

		cloned := typutil.DeepClone(node)

		if cloned == node {
			t.Errorf("cloned should be different pointer than original")
		}
		if cloned.Value != 1 {
			t.Errorf("cloned.Value should equal 1")
		}
		if cloned.Self != cloned {
			t.Errorf("cloned.Self should point to cloned (not original)")
		}
		if cloned.Self == node {
			t.Errorf("cloned.Self should not point to original node")
		}
	})

	t.Run("circular linked list", func(t *testing.T) {
		node1 := &selfRefNode{Value: 1}
		node2 := &selfRefNode{Value: 2}
		node3 := &selfRefNode{Value: 3}
		node1.Next = node2
		node2.Next = node3
		node3.Next = node1 // circular

		cloned := typutil.DeepClone(node1)

		// Verify structure
		if cloned.Value != 1 {
			t.Errorf("cloned.Value should equal 1")
		}
		if cloned.Next.Value != 2 {
			t.Errorf("cloned.Next.Value should equal 2")
		}
		if cloned.Next.Next.Value != 3 {
			t.Errorf("cloned.Next.Next.Value should equal 3")
		}
		if cloned.Next.Next.Next != cloned {
			t.Errorf("circular reference should point back to cloned")
		}

		// Verify independence
		if cloned == node1 || cloned.Next == node2 || cloned.Next.Next == node3 {
			t.Errorf("cloned nodes should be different from originals")
		}
	})
}

func TestDeepCloneSelfReferentialMap(t *testing.T) {
	t.Run("map containing itself", func(t *testing.T) {
		m := make(map[string]any)
		m["self"] = m
		m["value"] = 42

		cloned := typutil.DeepClone(m)

		if cloned["value"].(int) != 42 {
			t.Errorf("cloned[value] should equal 42")
		}

		// The "self" key should point to the cloned map, not the original
		selfRef, ok := cloned["self"].(map[string]any)
		if !ok {
			t.Errorf("cloned[self] should be map[string]any, got %T", cloned["self"])
		}

		// Verify it points to itself (the clone), not the original
		if selfRef["value"].(int) != 42 {
			t.Errorf("self-reference should have same value")
		}

		// Modify cloned and verify original is unaffected
		cloned["value"] = 100
		if m["value"].(int) != 42 {
			t.Errorf("modifying cloned should not affect original")
		}
	})
}

func TestDeepCloneSelfReferentialSlice(t *testing.T) {
	t.Run("slice containing pointer to itself via struct", func(t *testing.T) {
		type container struct {
			Items []*container
		}

		c := &container{}
		c.Items = []*container{c} // slice contains pointer to parent

		cloned := typutil.DeepClone(c)

		if cloned == c {
			t.Errorf("cloned should be different from original")
		}
		if len(cloned.Items) != 1 {
			t.Errorf("cloned.Items should have 1 element")
		}
		if cloned.Items[0] != cloned {
			t.Errorf("cloned.Items[0] should point to cloned, not original")
		}
	})
}
