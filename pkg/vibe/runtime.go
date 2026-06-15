package vibe

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// MotionPhase identifies which motion state should be resolved.
type MotionPhase string

const (
	MotionPhaseInitial MotionPhase = "initial"
	MotionPhaseAnimate MotionPhase = "animate"
	MotionPhaseExit    MotionPhase = "exit"
	MotionPhaseHover   MotionPhase = "hover"
	MotionPhaseTap     MotionPhase = "tap"
	MotionPhaseFocus   MotionPhase = "focus"
	MotionPhaseDrag    MotionPhase = "drag"
	MotionPhaseInView  MotionPhase = "inView"
)

// ReducedMotionMode describes how strongly motion should be reduced.
type ReducedMotionMode string

const (
	ReducedMotionAuto   ReducedMotionMode = "auto"
	ReducedMotionAlways ReducedMotionMode = "always"
	ReducedMotionNever  ReducedMotionMode = "never"
)

// MotionRuntimeConfig configures how a motion runtime should behave.
type MotionRuntimeConfig struct {
	ReducedMotion ReducedMotionMode `json:"reducedMotion,omitempty"`
	SystemReduced  bool             `json:"systemReduced,omitempty"`
}

// ShouldReduceMotion resolves the current reduced-motion policy.
func (cfg MotionRuntimeConfig) ShouldReduceMotion() bool {
	switch cfg.ReducedMotion {
	case ReducedMotionAlways:
		return true
	case ReducedMotionNever:
		return false
	default:
		return cfg.SystemReduced
	}
}

// MotionKeyframe describes a point on a motion timeline.
type MotionKeyframe struct {
	At         float64   `json:"at"`
	Style      StyleMap  `json:"style"`
	Transition Transition `json:"transition,omitempty"`
}

// MotionSequence is a timeline of motion keyframes.
type MotionSequence struct {
	Frames  []MotionKeyframe `json:"frames"`
	Loop    bool             `json:"loop,omitempty"`
	Reverse bool             `json:"reverse,omitempty"`
}

// Normalize sorts keyframes and fills safe defaults.
func (s MotionSequence) Normalize() MotionSequence {
	frames := make([]MotionKeyframe, 0, len(s.Frames))
	for _, frame := range s.Frames {
		if frame.Style == nil {
			frame.Style = StyleMap{}
		}
		frames = append(frames, frame)
	}

	sort.SliceStable(frames, func(i, j int) bool {
		if frames[i].At == frames[j].At {
			return i < j
		}
		return frames[i].At < frames[j].At
	})

	s.Frames = frames
	return s
}

// Sample resolves the style and transition at a normalized progress value.
func (s MotionSequence) Sample(progress float64) (StyleMap, Transition) {
	sequence := s.Normalize()
	if len(sequence.Frames) == 0 {
		return StyleMap{}, DefaultTransition()
	}

	if sequence.Loop {
		progress = math.Mod(progress, 1)
		if progress < 0 {
			progress += 1
		}
	}

	if sequence.Reverse {
		progress = 1 - clamp(progress, 0, 1)
	}

	if len(sequence.Frames) == 1 {
		return cloneStyleMap(sequence.Frames[0].Style), normalizeTransitionValue(sequence.Frames[0].Transition)
	}

	first := sequence.Frames[0]
	last := sequence.Frames[len(sequence.Frames)-1]
	if progress <= first.At {
		return cloneStyleMap(first.Style), normalizeTransitionValue(first.Transition)
	}
	if progress >= last.At {
		return cloneStyleMap(last.Style), normalizeTransitionValue(last.Transition)
	}

	for i := 1; i < len(sequence.Frames); i++ {
		right := sequence.Frames[i]
		if progress > right.At {
			continue
		}

		left := sequence.Frames[i-1]
		if right.At <= left.At {
			return cloneStyleMap(right.Style), mergeTransitionValues(left.Transition, right.Transition)
		}

		ratio := (progress - left.At) / (right.At - left.At)
		return interpolateStyleMaps(left.Style, right.Style, ratio), mergeTransitionValues(left.Transition, right.Transition)
	}

	return cloneStyleMap(last.Style), normalizeTransitionValue(last.Transition)
}

// StaggerConfig builds a predictable delay map for orchestrated motion.
type StaggerConfig struct {
	Count   int     `json:"count"`
	Delay   float64 `json:"delay,omitempty"`
	Step    float64 `json:"step,omitempty"`
	FromEnd bool    `json:"fromEnd,omitempty"`
}

// Delays returns a per-item stagger schedule.
func (c StaggerConfig) Delays() []float64 {
	if c.Count <= 0 {
		return nil
	}

	step := c.Step
	if step <= 0 {
		step = c.Delay
	}
	if step <= 0 {
		step = 0.1
	}

	delays := make([]float64, c.Count)
	for i := 0; i < c.Count; i++ {
		index := i
		if c.FromEnd {
			index = c.Count - 1 - i
		}
		delays[i] = c.Delay + (float64(index) * step)
	}
	return delays
}

// InertiaConfig projects motion with an easing tail.
type InertiaConfig struct {
	Velocity float64 `json:"velocity"`
	Friction float64 `json:"friction,omitempty"`
	Minimum  float64 `json:"minimum,omitempty"`
	Maximum  float64 `json:"maximum,omitempty"`
}

// ProjectInertia predicts the next value after a delta time.
func ProjectInertia(position float64, cfg InertiaConfig, dt float64) float64 {
	if dt <= 0 {
		return clamp(position, cfg.Minimum, cfg.Maximum)
	}

	friction := cfg.Friction
	if friction <= 0 {
		friction = 8
	}

	decay := math.Exp(-friction * dt)
	next := position + (cfg.Velocity * (1 - decay) / friction)
	return clamp(next, cfg.Minimum, cfg.Maximum)
}

// SpringState models a single spring step.
type SpringState struct {
	Position  float64 `json:"position"`
	Velocity  float64 `json:"velocity"`
	Target    float64 `json:"target"`
	Stiffness float64 `json:"stiffness,omitempty"`
	Damping   float64 `json:"damping,omitempty"`
	Mass      float64 `json:"mass,omitempty"`
}

// StepSpring advances a spring using a simple Euler step.
func StepSpring(state SpringState, dt float64) SpringState {
	if dt <= 0 {
		return state
	}

	stiffness := state.Stiffness
	if stiffness <= 0 {
		stiffness = 220
	}

	damping := state.Damping
	if damping <= 0 {
		damping = 22
	}

	mass := state.Mass
	if mass <= 0 {
		mass = 1
	}

	force := -stiffness * (state.Position - state.Target)
	resistive := -damping * state.Velocity
	acceleration := (force + resistive) / mass

	state.Velocity += acceleration * dt
	state.Position += state.Velocity * dt
	return state
}

// ApplyReducedMotion strips animation-heavy fields when motion should be reduced.
func ApplyReducedMotion(props MotionProps, cfg MotionRuntimeConfig) MotionProps {
	if !cfg.ShouldReduceMotion() {
		return props
	}

	if len(props.Animate) > 0 {
		props.Initial = MergeStyles(props.Initial, props.Animate)
	}

	props.Animate = nil
	props.Exit = nil
	props.WhileHover = nil
	props.WhileTap = nil
	props.WhileFocus = nil
	props.WhileDrag = nil
	props.WhileInView = nil
	props.Layout = false
	props.LayoutID = ""
	props.Transition = Transition{}
	return props
}

// ResolveMotionProps resolves a motion phase into a concrete style map and transition.
func ResolveMotionProps(props MotionProps, phase MotionPhase) (StyleMap, Transition, error) {
	baseStyle, baseTransition, err := resolveMotionPhaseBase(props, phase)
	if err != nil {
		return nil, Transition{}, err
	}

	if variant, ok := resolveMotionVariant(props, phase); ok {
		return MergeStyles(baseStyle, variant.Style), mergeTransitionValues(baseTransition, variant.Transition), nil
	}

	return baseStyle, baseTransition, nil
}

func resolveMotionPhaseBase(props MotionProps, phase MotionPhase) (StyleMap, Transition, error) {
	switch phase {
	case MotionPhaseInitial:
		return normalizeStyleMap(props.Initial), normalizeTransitionValue(props.Transition), nil
	case MotionPhaseAnimate:
		return MergeStyles(props.Initial, props.Animate), normalizeTransitionValue(props.Transition), nil
	case MotionPhaseExit:
		return MergeStyles(props.Animate, props.Exit), normalizeTransitionValue(props.Transition), nil
	case MotionPhaseHover:
		return MergeStyles(props.Animate, props.WhileHover), normalizeTransitionValue(props.Transition), nil
	case MotionPhaseTap:
		return MergeStyles(props.Animate, props.WhileTap), normalizeTransitionValue(props.Transition), nil
	case MotionPhaseFocus:
		return MergeStyles(props.Animate, props.WhileFocus), normalizeTransitionValue(props.Transition), nil
	case MotionPhaseDrag:
		return MergeStyles(props.Animate, props.WhileDrag), normalizeTransitionValue(props.Transition), nil
	case MotionPhaseInView:
		return MergeStyles(props.Animate, props.WhileInView), normalizeTransitionValue(props.Transition), nil
	default:
		return nil, Transition{}, fmt.Errorf("unknown motion phase %q", phase)
	}
}

func resolveMotionVariant(props MotionProps, phase MotionPhase) (Variant, bool) {
	if props.Variants == nil {
		props.Variants = VariantSet{}
	}

	if variant, ok := props.Variants[string(phase)]; ok {
		return variant, true
	}

	switch phase {
	case MotionPhaseHover:
		if name := strings.TrimSpace(props.Gestures.Hover); name != "" {
			variant, ok := props.Variants[name]
			return variant, ok
		}
	case MotionPhaseTap:
		if name := strings.TrimSpace(props.Gestures.Tap); name != "" {
			variant, ok := props.Variants[name]
			return variant, ok
		}
	case MotionPhaseFocus:
		if name := strings.TrimSpace(props.Gestures.Focus); name != "" {
			variant, ok := props.Variants[name]
			return variant, ok
		}
	case MotionPhaseDrag:
		if name := strings.TrimSpace(props.Gestures.Drag); name != "" {
			variant, ok := props.Variants[name]
			return variant, ok
		}
	case MotionPhaseInView:
		if name := strings.TrimSpace(props.Gestures.InView); name != "" {
			variant, ok := props.Variants[name]
			return variant, ok
		}
	}

	return Variant{}, false
}

func normalizeTransitionValue(t Transition) Transition {
	if t == (Transition{}) {
		return DefaultTransition()
	}
	return t.Normalize()
}

func mergeTransitionValues(base Transition, overlays ...Transition) Transition {
	out := base
	for _, overlay := range overlays {
		if overlay.Type != "" {
			out.Type = overlay.Type
		}
		if overlay.Duration != 0 {
			out.Duration = overlay.Duration
		}
		if overlay.Delay != 0 {
			out.Delay = overlay.Delay
		}
		if overlay.Ease != "" {
			out.Ease = overlay.Ease
		}
		if overlay.Stiffness != 0 {
			out.Stiffness = overlay.Stiffness
		}
		if overlay.Damping != 0 {
			out.Damping = overlay.Damping
		}
		if overlay.Mass != 0 {
			out.Mass = overlay.Mass
		}
		if overlay.Bounce != 0 {
			out.Bounce = overlay.Bounce
		}
		if overlay.Repeat != 0 {
			out.Repeat = overlay.Repeat
		}
		if overlay.RepeatType != "" {
			out.RepeatType = overlay.RepeatType
		}
	}

	return normalizeTransitionValue(out)
}

func interpolateStyleMaps(left, right StyleMap, ratio float64) StyleMap {
	result := StyleMap{}
	keys := map[string]struct{}{}

	for key := range left {
		keys[key] = struct{}{}
	}
	for key := range right {
		keys[key] = struct{}{}
	}

	for key := range keys {
		lv, lok := left[key]
		rv, rok := right[key]

		switch {
		case lok && rok:
			if leftFloat, ok := asFloat(lv); ok {
				if rightFloat, ok := asFloat(rv); ok {
					result[key] = lerp(leftFloat, rightFloat, ratio)
					continue
				}
			}

			if ratio < 0.5 {
				result[key] = lv
			} else {
				result[key] = rv
			}
		case lok:
			result[key] = lv
		case rok:
			result[key] = rv
		}
	}

	return result
}

func normalizeStyleMap(style StyleMap) StyleMap {
	if style == nil {
		return StyleMap{}
	}
	return cloneStyleMap(style)
}

func cloneStyleMap(style StyleMap) StyleMap {
	if style == nil {
		return StyleMap{}
	}

	clone := make(StyleMap, len(style))
	for key, value := range style {
		clone[key] = value
	}
	return clone
}

func clamp(value, minimum, maximum float64) float64 {
	if minimum != 0 && value < minimum {
		return minimum
	}
	if maximum != 0 && value > maximum {
		return maximum
	}
	return value
}

func lerp(start, end, ratio float64) float64 {
	return start + ((end - start) * ratio)
}
