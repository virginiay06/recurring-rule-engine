# recurring-rule-engine

Expands a compact recurrence rule into a series of time instances within a bounded range.

```go
package main

import (
	"fmt"
	"time"

	re "github.com/example/recurring-rule-engine"
)

func main() {
	start := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	r := re.Rule{
		Start:    start,
		Interval: 24 * time.Hour,
		Weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
	}

	from := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)

	occurrences, err := re.Expand(r, from, to)
	if err != nil {
		panic(err)
	}
	for _, t := range occurrences {
		fmt.Println(t)
	}
}
```

## Why this exists

Calendar recurrence is usually solved by pulling in an RFC 5545 library, but most
applications only need "every N units, optionally restricted to certain weekdays,
bounded by a start and an end". That small subset can be expressed without a parser
and without a dependency. The trade-off is that the rule is a Go struct rather than
a string: you lose human-readable rule text, and in exchange the library has no
lexer, no grammar, and no ambiguity about what a rule means.

## The awkward edge

The range passed to `Expand` is half-open: `from` is inclusive and `to` is exclusive.
This matches how Go's own `time` arithmetic is usually consumed, but it means that
an occurrence exactly at `to` is not returned, and an occurrence exactly at `from`
is. If you are paginating over a long recurrence, advance `from` to the last
returned occurrence plus one interval, not to the last returned occurrence itself.

A rule with an empty `Weekdays` slice allows every day. A rule with a non-empty
`Weekdays` slice allows only those days; the interval still advances from `Start`,
so the weekdays act as a filter rather than as a schedule of their own.
