package goscript

import (
	"errors"
	"strings"
	"testing"
)

func TestOptionAndResultChains(t *testing.T) {
	opt := Some("hello")
	if got := opt.AndThen(func(v interface{}) Option {
		return Some(v.(string) + " world")
	}).Expect("expected value"); got != "hello world" {
		t.Fatalf("unexpected option chain result: %v", got)
	}

	if none := None().OrElse(func() interface{} { return "fallback" }); none != "fallback" {
		t.Fatalf("unexpected fallback: %v", none)
	}

	res := ErrResult(errors.New("boom")).MapErr(func(err error) error {
		return errors.New("wrapped: " + err.Error())
	})

	if !res.IsErr() || !strings.Contains(res.Err.Error(), "wrapped:") {
		t.Fatalf("unexpected error result: %+v", res)
	}

	if got := Ok(10).AndThen(func(v interface{}) Result {
		return Ok(v.(int) + 5)
	}).Expect("expected result"); got != 15 {
		t.Fatalf("unexpected result chain output: %v", got)
	}
}

func TestMatchDetailed(t *testing.T) {
	result := MatchDetailed("admin",
		MatchCase{Equals: "guest", Then: func(v interface{}) interface{} { return "nope" }},
		MatchCase{Predicate: func(v interface{}) bool { return v == "admin" }, Then: func(v interface{}) interface{} { return "matched" }},
	)

	if !result.Matched || result.Index != 1 || result.Value != "matched" {
		t.Fatalf("unexpected match result: %+v", result)
	}
}
