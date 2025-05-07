/*
Task1 implements a naive solution to the Stim coding test.

usage:

	task1 [--date DATE] [--report N | --range A,B]

	where
		DATE is a naive date on the form YYYY-MM-DD, e.g 2025-05-05
		N is an integer in the range [1, 100]
		A,B is a pair of integers, both in the range [1, 100], with B >= A
*/
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// A PayoutEntry models a singular entry record in a payout report.
type PayoutEntry struct {
	Date   time.Time
	Name   string
	Amount uint64
}

// A PayoutReport contains zero or more payout entry records.
type PayoutReport struct {
	Entries []PayoutEntry
}

// Filter takes a predicate function `fn` and returns a new report that retains only
// those entries that satisfy the predicate.
func (report PayoutReport) filter(fn func(entry PayoutEntry) bool) PayoutReport {
	new_entries := make([]PayoutEntry, 0)

	for _, entry := range report.Entries {
		if fn(entry) {
			new_entries = append(new_entries, entry)
		}
	}

	return PayoutReport{new_entries}
}

// Sum computes the total sum of payout amounts in a report.
func (report PayoutReport) sum() uint64 {
	var total uint64

	for _, entry := range report.Entries {
		total += entry.Amount
	}

	return total
}

// A StrictCSVPayoutReportParser implements a strict parser for payout reports provided in
// CSV with fields 'date', 'name', 'amount' (strict ordering and field header required).
type StrictCSVPayoutReportParser struct{}

// TryParseUrl attempts to do a streaming parse of a CSV payout report fetched from a given url,
// either returning a PayoutReport or an error on failure.
func (parser StrictCSVPayoutReportParser) TryParseUrl(url string) (PayoutReport, error) {
	zeroResult := PayoutReport{}
	response, err := http.Get(url)

	if err != nil {
		return zeroResult, fmt.Errorf("HTTP request error: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return zeroResult, fmt.Errorf("Server returned non-200 status code")
	}

	defer response.Body.Close()

	return parser.TryParseStream(response.Body)
}

// TryParseStream attempts to do a streaming parse of a CSV payout report given in the form of
// an io.Reader stream, either returning a PayoutReport or an error on failure.
func (parser StrictCSVPayoutReportParser) TryParseStream(stream io.Reader) (PayoutReport, error) {
	zeroResult := PayoutReport{}
	result := PayoutReport{}

	reader := csv.NewReader(stream)

	header, err := reader.Read()

	if err != nil {
		return zeroResult, fmt.Errorf("CSV read error at line 1: %w", err)
	}

	if header[0] != "date" || header[1] != "name" || header[2] != "amount" {
		return zeroResult, fmt.Errorf("CSV header error")
	}

	nth_line := 1

	for {
		nth_line += 1
		fields, err := reader.Read()

		if err != nil && err.Error() == "EOF" {
			break
		}

		if err != nil {
			return zeroResult, fmt.Errorf("CSV read error at line %d: %w", nth_line, err)
		}

		date, err := time.Parse("2006-01-02", fields[0])

		if err != nil {
			return zeroResult, fmt.Errorf("CSV parse error, invalid date at line %d: %w", nth_line, err)
		}

		if fields[1] == "" {
			return zeroResult, fmt.Errorf("CSV parse error, empty name at line %d", nth_line)
		}

		amount, err := strconv.ParseUint(fields[2], 10, 64)

		if err != nil {
			return zeroResult, fmt.Errorf("CSV parse error, invalid amount at line %d: %w", nth_line, err)
		}

		result.Entries = append(result.Entries, PayoutEntry{
			date,
			fields[1],
			amount,
		})
	}

	return result, nil
}

func main() {
	flag_target_date := flag.String("date", time.Now().Format("2006-01-02"), "YYYY-MM-DD. Default: current date.")
	flag_report := flag.String("report", "", "Integer in the range [1, 100]. Overrides -range.")
	flag_range := flag.String("range", "1,100", "Comma-separated pair of integers, e.g 1,10. Endpoints in range [1, 100] and end >= start.")

	flag.Parse()

	if *flag_report != "" {
		*flag_range = ""
	}

	target_date, err := time.Parse("2006-01-02", *flag_target_date)

	if err != nil {
		log.Fatalln("FATAL: Invalid date. Expected format is YYYY-MM-DD.")
	}

	var target_range_start int
	var target_range_end int

	if *flag_report != "" {
		target_report, err := strconv.ParseInt(*flag_report, 10, 32)

		if err != nil || target_report < 1 || target_report > 100 {
			log.Fatalln("FATAL: Invalid target report. Expected integer in range [1, 100].")
		}

		target_range_start = int(target_report)
		target_range_end = int(target_report)
	}

	if *flag_range != "" {
		fields := strings.Split(*flag_range, ",")

		if len(fields) != 2 {
			log.Fatalln("FATAL: Invalid range. Expected pair of integers, e.g 1,10.")
		}

		target_start, err := strconv.ParseInt(fields[0], 10, 32)

		if err != nil || target_start < 1 || target_start > 100 {
			log.Fatalln("FATAL: Invalid range start. Expected integer in range [1, 100].")
		}

		target_end, err := strconv.ParseInt(fields[1], 10, 32)

		if err != nil || target_end < 1 || target_end > 100 || target_end < target_start {
			log.Fatalf("FATAL: Invalid range end. Expected integer in range [%d, 100].", target_start)
		}

		target_range_start = int(target_start)
		target_range_end = int(target_end)
	}

	parser := StrictCSVPayoutReportParser{}
	var total uint64

	for n := range target_range_end - target_range_start + 1 {
		report_num := n + target_range_start
		report, err := parser.TryParseUrl(fmt.Sprintf("https://codetest.stim.se/payouts/%d", report_num))

		if err != nil {
			log.Fatalf("FATAL: Report fetch/parse error: %s\n", err)
		}

		total += report.filter(func(entry PayoutEntry) bool {
			return entry.Date == target_date
		}).sum()
	}

	fmt.Printf(
		"{ \n"+
			"  \"date\": \"%s\",\n"+
			"  \"reportRangeStart\": %d,\n"+
			"  \"reportRangeEnd\": %d,\n"+
			"  \"totalPayout\": %d\n"+
			"}\n",
		target_date.Format("2006-01-02"),
		target_range_start,
		target_range_end,
		total,
	)
}
