package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func unwrap[T any](value T, e error) T {
	if e != nil {
		panic(e)
	} else {
		return value
	}
}

func assertEqual[T comparable](t *testing.T, left T, right T) {
	if left != right {
		t.Errorf("Assertion failed: %s != %s\n", fmt.Sprint(left), fmt.Sprint(right))
	}
}

func assertError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Assertion failed: err is nil\n")
	}
}

func assertDeepEqual[T any](t *testing.T, left T, right T) {
	if !reflect.DeepEqual(left, right) {
		t.Errorf("Assertion failed: %s != %s", fmt.Sprint(left), fmt.Sprint(right))
	}
}

func exampleReport1() PayoutReport {
	return PayoutReport{[]PayoutEntry{
		{
			unwrap(time.Parse("2006-01-02", "2025-05-03")),
			"Alice",
			200,
		},
		{
			unwrap(time.Parse("2006-01-02", "2025-05-04")),
			"Bob",
			300,
		},
		{
			unwrap(time.Parse("2006-01-02", "2025-05-04")),
			"Cheyenne",
			500,
		},
		{
			unwrap(time.Parse("2006-01-02", "2025-05-05")),
			"Douglas",
			700,
		},
	}}
}

func exampleReport2() PayoutReport {
	return PayoutReport{[]PayoutEntry{
		{
			unwrap(time.Parse("2006-01-02", "2025-05-03")),
			"Alice",
			200,
		},
		{
			unwrap(time.Parse("2006-01-02", "2025-05-04")),
			"Bob",
			300,
		},
	}}
}

func TestPayoutReportFilter(t *testing.T) {
	report := exampleReport1()

	filtered0503 := report.filter(func(entry PayoutEntry) bool {
		return entry.Date == unwrap(time.Parse("2006-01-02", "2025-05-03"))
	})

	filtered0504 := report.filter(func(entry PayoutEntry) bool {
		return entry.Date == unwrap(time.Parse("2006-01-02", "2025-05-04"))
	})

	filteredAlice := report.filter(func(entry PayoutEntry) bool {
		return entry.Name == "Alice"
	})

	filteredGte300 := report.filter(func(entry PayoutEntry) bool {
		return entry.Amount >= 300
	})

	assertEqual(t, len(filtered0503.Entries), 1)
	assertEqual(t, len(filtered0504.Entries), 2)
	assertEqual(t, len(filteredAlice.Entries), 1)
	assertEqual(t, len(filteredGte300.Entries), 3)

	assertDeepEqual(t, filtered0503, filteredAlice)
}

func TestPayoutReportSum(t *testing.T) {
	assertEqual(t, exampleReport1().sum(), 1700)
	assertEqual(t, exampleReport2().sum(), 500)
}

func TestParseFirstReport(t *testing.T) {
	parser := StrictCSVPayoutReportParser{}
	report := unwrap(parser.TryParseStream(unwrap(os.Open("./1.csv"))))

	assertEqual(t, len(report.Entries), 119925)
	assertEqual(t, report.sum(), 90101657)
}

func TestParseInvalid(t *testing.T) {
	parser := StrictCSVPayoutReportParser{}

	_, err := parser.TryParseStream(strings.NewReader(""))
	assertError(t, err)

	_, err = parser.TryParseStream(strings.NewReader("2002-03-04,Alice,123\n"))
	assertError(t, err)

	_, err = parser.TryParseStream(strings.NewReader("date,name,amount\n123,Alice,123\n"))
	assertError(t, err)

	_, err = parser.TryParseStream(strings.NewReader("date,name,amount\n2002-03-04,,123\n"))
	assertError(t, err)

	_, err = parser.TryParseStream(strings.NewReader("date,name,amount\n2002-03-04,Alice,-123\n"))
	assertError(t, err)

	_, err = parser.TryParseStream(strings.NewReader("date,name,amount\n2002-03-04,Alice,123,foo\n"))
	assertError(t, err)
}
