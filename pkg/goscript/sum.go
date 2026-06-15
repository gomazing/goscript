package goscript

import (
	"fmt"
	"strings"
)

// Variant models a tagged value that can participate in sum-type style matching.
type Variant struct {
	TypeName string
	Tag      string
	Value    interface{}
}

// NewVariant creates a tagged value for a named sum type.
func NewVariant(typeName, tag string, value interface{}) Variant {
	return Variant{
		TypeName: strings.TrimSpace(typeName),
		Tag:      strings.TrimSpace(tag),
		Value:    value,
	}
}

// Normalize returns a stable copy of the variant with trimmed identifiers.
func (v Variant) Normalize() Variant {
	v.TypeName = strings.TrimSpace(v.TypeName)
	v.Tag = strings.TrimSpace(v.Tag)
	return v
}

// IsType reports whether the variant belongs to the supplied sum type.
func (v Variant) IsType(typeName string) bool {
	return strings.TrimSpace(v.TypeName) != "" && strings.TrimSpace(v.TypeName) == strings.TrimSpace(typeName)
}

// Is reports whether the variant uses the supplied tag.
func (v Variant) Is(tag string) bool {
	return strings.TrimSpace(v.Tag) != "" && strings.TrimSpace(v.Tag) == strings.TrimSpace(tag)
}

// Unwrap returns the variant metadata and payload.
func (v Variant) Unwrap() (string, string, interface{}) {
	v = v.Normalize()
	return v.TypeName, v.Tag, v.Value
}

// Expect returns the payload or panics if the type or tag does not match.
func (v Variant) Expect(typeName, tag string) interface{} {
	if !v.IsType(typeName) || !v.Is(tag) {
		panic(fmt.Sprintf("expected %s.%s, got %s.%s", strings.TrimSpace(typeName), strings.TrimSpace(tag), v.TypeName, v.Tag))
	}
	return v.Value
}

// String returns a debug-friendly representation of the variant.
func (v Variant) String() string {
	v = v.Normalize()
	if v.TypeName == "" {
		return fmt.Sprintf("%s(%v)", v.Tag, v.Value)
	}
	return fmt.Sprintf("%s.%s(%v)", v.TypeName, v.Tag, v.Value)
}

// VariantCase describes one exhaustive branch for a tagged value.
type VariantCase struct {
	Tag  string
	Then func(interface{}) interface{}
}

// VariantSpec declares the full set of legal tags for a named sum type.
type VariantSpec struct {
	Name string
	Tags []string
}

// NewVariantSpec creates a normalized spec for a sum type.
func NewVariantSpec(name string, tags ...string) VariantSpec {
	spec := VariantSpec{Name: strings.TrimSpace(name), Tags: append([]string(nil), tags...)}
	return spec.Normalize()
}

// Normalize trims names and removes duplicate tags while preserving order.
func (s VariantSpec) Normalize() VariantSpec {
	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		s.Name = "variant"
	}

	seen := make(map[string]struct{}, len(s.Tags))
	normalized := make([]string, 0, len(s.Tags))
	for _, tag := range s.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	s.Tags = normalized
	return s
}

// Validate ensures the spec has enough information to support exhaustive matching.
func (s VariantSpec) Validate() error {
	s = s.Normalize()
	if len(s.Tags) == 0 {
		return fmt.Errorf("variant spec %q must define at least one tag", s.Name)
	}
	return nil
}

// Has reports whether the spec contains a tag.
func (s VariantSpec) Has(tag string) bool {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return false
	}
	s = s.Normalize()
	for _, candidate := range s.Tags {
		if candidate == tag {
			return true
		}
	}
	return false
}

// MissingCases returns the tags declared by the spec that were not covered by cases.
func (s VariantSpec) MissingCases(cases ...VariantCase) []string {
	s = s.Normalize()
	provided := make(map[string]struct{}, len(cases))
	for _, c := range cases {
		tag := strings.TrimSpace(c.Tag)
		if tag == "" {
			continue
		}
		provided[tag] = struct{}{}
	}

	missing := make([]string, 0)
	for _, tag := range s.Tags {
		if _, ok := provided[tag]; !ok {
			missing = append(missing, tag)
		}
	}
	return missing
}

// UnknownCases returns cases that are not part of the spec.
func (s VariantSpec) UnknownCases(cases ...VariantCase) []string {
	s = s.Normalize()
	unknown := make([]string, 0)
	for _, c := range cases {
		tag := strings.TrimSpace(c.Tag)
		if tag == "" {
			continue
		}
		if !s.Has(tag) {
			unknown = append(unknown, tag)
		}
	}
	return unknown
}

// ExhaustiveMatch evaluates a tagged value and guarantees coverage for the whole spec.
func (s VariantSpec) ExhaustiveMatch(value Variant, cases ...VariantCase) (interface{}, error) {
	s = s.Normalize()
	if err := s.Validate(); err != nil {
		return nil, err
	}

	value = value.Normalize()
	if value.TypeName != "" && value.TypeName != s.Name {
		return nil, fmt.Errorf("variant type mismatch: expected %q, got %q", s.Name, value.TypeName)
	}
	if !s.Has(value.Tag) {
		return nil, fmt.Errorf("variant tag %q is not part of spec %q", value.Tag, s.Name)
	}

	caseMap := make(map[string]VariantCase, len(cases))
	for _, c := range cases {
		tag := strings.TrimSpace(c.Tag)
		if tag == "" {
			continue
		}
		if _, exists := caseMap[tag]; exists {
			return nil, fmt.Errorf("duplicate case for variant tag %q in spec %q", tag, s.Name)
		}
		caseMap[tag] = VariantCase{Tag: tag, Then: c.Then}
	}

	if missing := s.MissingCases(cases...); len(missing) > 0 {
		return nil, fmt.Errorf("non-exhaustive match for %q: missing cases for %s", s.Name, strings.Join(missing, ", "))
	}
	if unknown := s.UnknownCases(cases...); len(unknown) > 0 {
		return nil, fmt.Errorf("match for %q includes unknown cases: %s", s.Name, strings.Join(unknown, ", "))
	}

	selected, ok := caseMap[value.Tag]
	if !ok {
		return nil, fmt.Errorf("variant tag %q missing from match for %q", value.Tag, s.Name)
	}
	if selected.Then != nil {
		return selected.Then(value.Value), nil
	}
	return value.Value, nil
}

// MustMatch is a convenience helper for exhaustive matches that should never fail.
func (s VariantSpec) MustMatch(value Variant, cases ...VariantCase) interface{} {
	out, err := s.ExhaustiveMatch(value, cases...)
	if err != nil {
		panic(err)
	}
	return out
}
