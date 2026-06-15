package goscript

import "fmt"

// Option models an optional value without forcing nil checks everywhere.
type Option struct {
	Value interface{}
	Ok    bool
}

// Some creates a populated option.
func Some(value interface{}) Option {
	return Option{Value: value, Ok: true}
}

// None creates an empty option.
func None() Option {
	return Option{Ok: false}
}

// IsSome reports whether the option contains a value.
func (o Option) IsSome() bool {
	return o.Ok
}

// IsNone reports whether the option is empty.
func (o Option) IsNone() bool {
	return !o.Ok
}

// UnwrapOr returns a fallback when the option is empty.
func (o Option) UnwrapOr(fallback interface{}) interface{} {
	if o.Ok {
		return o.Value
	}
	return fallback
}

// OrElse computes a fallback when the option is empty.
func (o Option) OrElse(fallback func() interface{}) interface{} {
	if o.Ok {
		return o.Value
	}
	if fallback == nil {
		return nil
	}
	return fallback()
}

// Expect returns the contained value or panics with the supplied message.
func (o Option) Expect(message string) interface{} {
	if !o.Ok {
		panic(message)
	}
	return o.Value
}

// Map transforms the contained value if present.
func (o Option) Map(transform func(interface{}) interface{}) Option {
	if !o.Ok || transform == nil {
		return o
	}
	return Some(transform(o.Value))
}

// AndThen chains another optional computation.
func (o Option) AndThen(transform func(interface{}) Option) Option {
	if !o.Ok || transform == nil {
		return o
	}
	return transform(o.Value)
}

// Filter keeps the option only when the predicate matches.
func (o Option) Filter(predicate func(interface{}) bool) Option {
	if !o.Ok || predicate == nil {
		return o
	}
	if predicate(o.Value) {
		return o
	}
	return None()
}

// Inspect runs a side-effect when the option is populated.
func (o Option) Inspect(fn func(interface{})) Option {
	if o.Ok && fn != nil {
		fn(o.Value)
	}
	return o
}

// Result models a success or failure outcome.
type Result struct {
	Value interface{}
	Err   error
}

// Ok creates a successful result.
func Ok(value interface{}) Result {
	return Result{Value: value}
}

// ErrResult creates a failed result.
func ErrResult(err error) Result {
	return Result{Err: err}
}

// Unwrap returns the success value and error pair.
func (r Result) Unwrap() (interface{}, error) {
	return r.Value, r.Err
}

// IsOk reports whether the result succeeded.
func (r Result) IsOk() bool {
	return r.Err == nil
}

// IsErr reports whether the result failed.
func (r Result) IsErr() bool {
	return r.Err != nil
}

// UnwrapOr returns the value or a fallback when failed.
func (r Result) UnwrapOr(fallback interface{}) interface{} {
	if r.Err != nil {
		return fallback
	}
	return r.Value
}

// OrElse computes a replacement value when the result fails.
func (r Result) OrElse(fallback func(error) interface{}) Result {
	if r.Err == nil {
		return r
	}
	if fallback == nil {
		return Ok(nil)
	}
	return Ok(fallback(r.Err))
}

// Expect returns the value or panics with context.
func (r Result) Expect(message string) interface{} {
	if r.Err != nil {
		panic(fmt.Sprintf("%s: %v", message, r.Err))
	}
	return r.Value
}

// Map transforms the value when the result succeeded.
func (r Result) Map(transform func(interface{}) interface{}) Result {
	if r.Err != nil || transform == nil {
		return r
	}
	return Ok(transform(r.Value))
}

// MapErr transforms the error when the result failed.
func (r Result) MapErr(transform func(error) error) Result {
	if r.Err == nil || transform == nil {
		return r
	}
	return ErrResult(transform(r.Err))
}

// AndThen chains another result-producing computation.
func (r Result) AndThen(transform func(interface{}) Result) Result {
	if r.Err != nil || transform == nil {
		return r
	}
	return transform(r.Value)
}

// Recover converts a failed result into a successful value.
func (r Result) Recover(transform func(error) interface{}) Result {
	if r.Err == nil {
		return r
	}
	if transform == nil {
		return r
	}
	return Ok(transform(r.Err))
}

// Inspect runs a side effect on successful values.
func (r Result) Inspect(fn func(interface{})) Result {
	if r.Err == nil && fn != nil {
		fn(r.Value)
	}
	return r
}
