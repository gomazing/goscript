package goscript

import "reflect"

// MatchCase represents one branch in a pattern match.
type MatchCase struct {
	Equals    interface{}
	Kind      reflect.Kind
	Predicate func(interface{}) bool
	Then      func(interface{}) interface{}
}

// MatchResult captures whether a value matched and which branch won.
type MatchResult struct {
	Matched bool
	Index   int
	Value   interface{}
}

// OrElse returns the match value or a fallback when no case matched.
func (r MatchResult) OrElse(fallback interface{}) interface{} {
	if r.Matched {
		return r.Value
	}
	return fallback
}

// MatchDetailed evaluates the first branch that fits the value and reports the branch index.
func MatchDetailed(value interface{}, cases ...MatchCase) MatchResult {
	for idx, c := range cases {
		if c.Predicate != nil && c.Predicate(value) {
			if c.Then != nil {
				return MatchResult{Matched: true, Index: idx, Value: c.Then(value)}
			}
			return MatchResult{Matched: true, Index: idx, Value: value}
		}

		if c.Equals != nil && reflect.DeepEqual(value, c.Equals) {
			if c.Then != nil {
				return MatchResult{Matched: true, Index: idx, Value: c.Then(value)}
			}
			return MatchResult{Matched: true, Index: idx, Value: value}
		}

		if c.Kind != reflect.Invalid {
			if value != nil && reflect.TypeOf(value).Kind() == c.Kind {
				if c.Then != nil {
					return MatchResult{Matched: true, Index: idx, Value: c.Then(value)}
				}
				return MatchResult{Matched: true, Index: idx, Value: value}
			}
		}
	}

	return MatchResult{}
}

// Match evaluates the first branch that fits the value.
func Match(value interface{}, cases ...MatchCase) interface{} {
	return MatchDetailed(value, cases...).Value
}

// MatchWithDefault returns the first match or the provided fallback when no case matches.
func MatchWithDefault(value interface{}, fallback interface{}, cases ...MatchCase) interface{} {
	result := MatchDetailed(value, cases...)
	if result.Matched {
		return result.Value
	}
	return fallback
}
