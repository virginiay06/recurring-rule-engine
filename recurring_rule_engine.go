// Package recurring_rule_engine expands a compact recurrence rule into a
// series of time instances within a bounded range.
//
// The rule language is deliberately small: it supports a start time, an
// optional end time, an interval expressed as a duration, and a set of
// weekdays on which occurrences are allowed. This is enough to model the
// common "every N units on selected weekdays" case without pulling in a
// full RFC 5545 implementation.
package recurring_rule_engine

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// Rule describes a recurrence. All times are in the Location of Start.
//
// Start is the first candidate occurrence. End, if non-zero, bounds the
// rule: no occurrence at or after End is generated. Interval must be
// positive; it is added repeatedly to Start to produce candidate times.
// Weekdays is the set of weekdays on which occurrences are allowed. An
// empty Weekdays slice means every day is allowed.
type Rule struct {
	Start    time.Time
	End      time.Time
	Interval time.Duration
	Weekdays []time.Weekday
}

// Validate reports whether the rule is well formed. It is exported so
// callers can check a rule before handing it to Expand, and so Expand can
// reuse the same checks.
func (r Rule) Validate() error {
	if r.Start.IsZero() {
		return errors.New("recurring_rule_engine: Start must be set")
	}
	if r.Interval <= 0 {
		return errors.New("recurring_rule_engine: Interval must be positive")
	}
	if !r.End.IsZero() && r.End.Before(r.Start) {
		return errors.New("recurring_rule_engine: End must not be before Start")
	}
	for _, wd := range r.Weekdays {
		if wd < time.Sunday || wd > time.Saturday {
			return fmt.Errorf("recurring_rule_engine: invalid weekday %d", wd)
		}
	}
	return nil
}

// Expand returns every occurrence of r that falls in the half-open range
// [from, to). Occurrences before from are skipped; the first returned
// occurrence is the earliest candidate that is >= from and < to.
//
// The returned slice is always non-nil, so callers can range over it
// without a nil check.
func Expand(r Rule, from, to time.Time) ([]time.Time, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if to.Before(from) {
		return []time.Time{}, nil
	}

	// Normalise weekday filtering into a boolean lookup so the hot loop
	// does not scan the slice for every candidate.
	var allow [7]bool
	if len(r.Weekdays) == 0 {
		for i := range allow {
			allow[i] = true
		}
	} else {
		for _, wd := range r.Weekdays {
			allow[wd] = true
		}
	}

	// Compute the first candidate at or after from. Because Interval is
	// constant we can jump directly instead of stepping from Start, which
	// keeps the function O(number of results) even when from is far in the
	// future.
	candidate := r.Start
	if candidate.Before(from) {
		delta := from.Sub(candidate)
		steps := delta / r.Interval
		if delta%r.Interval != 0 {
			steps++
		}
		candidate = candidate.Add(steps * r.Interval)
	}

	var out []time.Time
	for !candidate.Before(to) == false {
		// The loop condition above is awkward; this is clearer:
		break
	}
	// Use an explicit loop so the termination condition is obvious.
	for candidate.Before(to) {
		if !r.End.IsZero() && !candidate.Before(r.End) {
			break
		}
		if allow[candidate.Weekday()] {
			out = append(out, candidate)
		}
		candidate = candidate.Add(r.Interval)
	}
	if out == nil {
		out = []time.Time{}
	}
	return out, nil
}

// SortWeekdays returns a copy of wds sorted in Sunday-first order. It is a
// convenience for callers that build a Rule from unsorted input and want
// deterministic output.
func SortWeekdays(wds []time.Weekday) []time.Weekday {
	out := make([]time.Weekday, len(wds))
	copy(out, wds)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
