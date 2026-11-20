package recurring_rule_engine

import (
	"testing"
	"time"
)

func mustTime(t *testing.T, layout, value string) time.Time {
	t.Helper()
	tm, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return tm
}

func TestExpandDaily(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-04T00:00:00Z")

	r := Rule{Start: start, Interval: 24 * time.Hour}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	want := []time.Time{
		mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z"),
		mustTime(t, time.RFC3339, "2024-01-02T00:00:00Z"),
		mustTime(t, time.RFC3339, "2024-01-03T00:00:00Z"),
	}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("got[%d]=%v want %v", i, got[i], want[i])
		}
	}
}

func TestExpandStartsAfterFrom(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-01T12:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-03T00:00:00Z")

	r := Rule{Start: start, Interval: 24 * time.Hour}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	want := []time.Time{
		mustTime(t, time.RFC3339, "2024-01-02T00:00:00Z"),
	}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d: %v", len(got), len(want), got)
	}
	if !got[0].Equal(want[0]) {
		t.Errorf("got[0]=%v want %v", got[0], want[0])
	}
}

func TestExpandWeekdayFilter(t *testing.T) {
	// 2024-01-01 is a Monday. Allow only Wednesdays.
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-11T00:00:00Z")

	r := Rule{Start: start, Interval: 24 * time.Hour, Weekdays: []time.Weekday{time.Wednesday}}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	want := []time.Time{
		mustTime(t, time.RFC3339, "2024-01-03T00:00:00Z"),
		mustTime(t, time.RFC3339, "2024-01-10T00:00:00Z"),
	}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("got[%d]=%v want %v", i, got[i], want[i])
		}
	}
}

func TestExpandEndBound(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	end := mustTime(t, time.RFC3339, "2024-01-03T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-10T00:00:00Z")

	r := Rule{Start: start, End: end, Interval: 24 * time.Hour}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	want := []time.Time{
		mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z"),
		mustTime(t, time.RFC3339, "2024-01-02T00:00:00Z"),
	}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("got[%d]=%v want %v", i, got[i], want[i])
		}
	}
}

func TestExpandEmptyRange(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-05T00:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")

	r := Rule{Start: start, Interval: 24 * time.Hour}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if got == nil {
		t.Fatal("got nil slice, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestExpandNoCandidatesInRange(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-02T00:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-03T00:00:00Z")

	r := Rule{Start: start, Interval: 24 * time.Hour, Weekdays: []time.Weekday{time.Wednesday}}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestExpandToIsExclusive(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-02T00:00:00Z")

	r := Rule{Start: start, Interval: 24 * time.Hour}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %v, want exactly one occurrence", got)
	}
}

func TestExpandFromIsInclusive(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	from := mustTime(t, time.RFC3339, "2024-01-02T00:00:00Z")
	to := mustTime(t, time.RFC3339, "2024-01-03T00:00:00Z")

	r := Rule{Start: start, Interval: 24 * time.Hour}
	got, err := Expand(r, from, to)
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %v, want exactly one occurrence", got)
	}
	if !got[0].Equal(from) {
		t.Errorf("got[0]=%v want %v", got[0], from)
	}
}

func TestValidateRejectsBadRules(t *testing.T) {
	start := mustTime(t, time.RFC3339, "2024-01-01T00:00:00Z")
	end := mustTime(t, time.RFC3339, "2023-12-31T00:00:00Z")

	cases := []struct {
		name string
		r    Rule
	}{
		{"zero start", Rule{Interval: time.Hour}},
		{"zero interval", Rule{Start: start}},
		{"negative interval", Rule{Start: start, Interval: -time.Hour}},
		{"end before start", Rule{Start: start, End: end, Interval: time.Hour}},
		{"invalid weekday", Rule{Start: start, Interval: time.Hour, Weekdays: []time.Weekday{time.Weekday(9)}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.r.Validate(); err == nil {
				t.Fatal("Validate returned nil, want error")
			}
		})
	}
}

func TestExpandPropagatesValidationError(t *testing.T) {
	_, err := Expand(Rule{}, time.Now(), time.Now().Add(time.Hour))
	if err == nil {
		t.Fatal("Expand returned nil error, want validation error")
	}
}

func TestSortWeekdays(t *testing.T) {
	in := []time.Weekday{time.Friday, time.Monday, time.Wednesday}
	got := SortWeekdays(in)
	want := []time.Weekday{time.Monday, time.Wednesday, time.Friday}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d]=%v want %v", i, got[i], want[i])
		}
	}
	// The input slice must not be mutated.
	if in[0] != time.Friday {
		t.Errorf("input mutated: %v", in)
	}
}
