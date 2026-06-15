package goscript

import (
	"strings"
	"testing"
)

func TestVariantExhaustiveMatch(t *testing.T) {
	spec := NewVariantSpec("Maybe", "Some", "None")
	value := NewVariant("Maybe", "Some", "hello")

	got, err := spec.ExhaustiveMatch(value,
		VariantCase{Tag: "Some", Then: func(v interface{}) interface{} {
			return v.(string) + " world"
		}},
		VariantCase{Tag: "None", Then: func(v interface{}) interface{} {
			return "missing"
		}},
	)
	if err != nil {
		t.Fatalf("ExhaustiveMatch returned error: %v", err)
	}

	if got != "hello world" {
		t.Fatalf("unexpected variant match result: %v", got)
	}
}

func TestVariantExhaustiveMatchMissingCase(t *testing.T) {
	spec := NewVariantSpec("Maybe", "Some", "None")
	value := NewVariant("Maybe", "Some", "hello")

	_, err := spec.ExhaustiveMatch(value,
		VariantCase{Tag: "Some", Then: func(v interface{}) interface{} {
			return v
		}},
	)
	if err == nil {
		t.Fatalf("expected error for non-exhaustive match")
	}

	if !strings.Contains(err.Error(), "missing cases for None") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVariantExhaustiveMatchUnknownTag(t *testing.T) {
	spec := NewVariantSpec("Maybe", "Some", "None")
	value := NewVariant("Maybe", "MaybeMaybe", "hello")

	_, err := spec.ExhaustiveMatch(value,
		VariantCase{Tag: "Some", Then: func(v interface{}) interface{} { return v }},
		VariantCase{Tag: "None", Then: func(v interface{}) interface{} { return v }},
	)
	if err == nil {
		t.Fatalf("expected error for unknown tag")
	}

	if !strings.Contains(err.Error(), "is not part of spec") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVariantSpecMustMatch(t *testing.T) {
	spec := NewVariantSpec("Either", "Left", "Right")
	got := spec.MustMatch(NewVariant("Either", "Right", 42),
		VariantCase{Tag: "Left", Then: func(v interface{}) interface{} { return v }},
		VariantCase{Tag: "Right", Then: func(v interface{}) interface{} {
			return v.(int) + 1
		}},
	)

	if got != 43 {
		t.Fatalf("unexpected MustMatch result: %v", got)
	}
}
