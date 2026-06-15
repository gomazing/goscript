package vibe

import (
	"math"
	"testing"
)

func TestResolveMotionPropsGestureVariant(t *testing.T) {
	props := MotionProps{
		Animate: StyleMap{"opacity": 1},
		Variants: VariantSet{
			"lift": {
				Name:  "lift",
				Style: StyleMap{"scale": 1.1},
				Transition: Transition{
					Type:     TransitionTween,
					Duration: 0.2,
				},
			},
		},
		Gestures: GestureTargets{
			Hover: "lift",
		},
	}

	style, transition, err := ResolveMotionProps(props, MotionPhaseHover)
	if err != nil {
		t.Fatalf("ResolveMotionProps returned error: %v", err)
	}

	if got := style["scale"]; got != 1.1 {
		t.Fatalf("expected scale 1.1, got %v", got)
	}

	if transition.Type != TransitionTween || transition.Duration != 0.2 {
		t.Fatalf("unexpected transition: %+v", transition)
	}
}

func TestMotionSequenceSampleInterpolates(t *testing.T) {
	sequence := MotionSequence{
		Frames: []MotionKeyframe{
			{At: 0, Style: StyleMap{"opacity": 0, "x": 0}},
			{At: 1, Style: StyleMap{"opacity": 1, "x": 100}},
		},
	}

	style, transition := sequence.Sample(0.5)

	if diff := math.Abs(style["opacity"].(float64) - 0.5); diff > 0.0001 {
		t.Fatalf("expected opacity near 0.5, got %v", style["opacity"])
	}
	if diff := math.Abs(style["x"].(float64) - 50); diff > 0.0001 {
		t.Fatalf("expected x near 50, got %v", style["x"])
	}
	if transition.Type != TransitionSpring {
		t.Fatalf("expected default transition type spring, got %s", transition.Type)
	}
}

func TestReducedMotionCollapsesAnimation(t *testing.T) {
	props := MotionProps{
		Initial: StyleMap{"opacity": 0},
		Animate: StyleMap{"opacity": 1},
		Exit:    StyleMap{"opacity": 0},
		Layout:  true,
		Transition: Transition{
			Type:     TransitionSpring,
			Duration: 0.4,
		},
	}

	reduced := ApplyReducedMotion(props, MotionRuntimeConfig{ReducedMotion: ReducedMotionAlways})

	if reduced.Layout {
		t.Fatalf("expected layout motion to be disabled")
	}
	if reduced.Transition != (Transition{}) {
		t.Fatalf("expected transition to be cleared, got %+v", reduced.Transition)
	}
	if reduced.Initial["opacity"] != 1 {
		t.Fatalf("expected initial opacity to collapse to animate state, got %v", reduced.Initial["opacity"])
	}
	if reduced.Animate != nil || reduced.Exit != nil {
		t.Fatalf("expected animation surfaces to be cleared: %+v", reduced)
	}
}

func TestSequenceHelpers(t *testing.T) {
	delays := StaggerConfig{Count: 3, Delay: 0.1, Step: 0.2}.Delays()
	if len(delays) != 3 {
		t.Fatalf("expected 3 delays, got %d", len(delays))
	}
	if diff := math.Abs(delays[0] - 0.1); diff > 0.0001 {
		t.Fatalf("unexpected first delay: %v", delays[0])
	}
	if diff := math.Abs(delays[2] - 0.5); diff > 0.0001 {
		t.Fatalf("unexpected last delay: %v", delays[2])
	}

	next := ProjectInertia(10, InertiaConfig{Velocity: 4, Friction: 6, Minimum: 0, Maximum: 20}, 0.5)
	if next <= 10 {
		t.Fatalf("expected inertia projection to move forward, got %v", next)
	}

	spring := StepSpring(SpringState{Position: 0, Velocity: 0, Target: 1}, 0.016)
	if spring.Position <= 0 {
		t.Fatalf("expected spring position to advance, got %+v", spring)
	}
}
