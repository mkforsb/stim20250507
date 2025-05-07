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

type PayoutEntry struct {
	Date   time.Time
	Name   string
	Amount uint64
}

type StrictCSVPayoutReportSummarizer struct{}

func (parser StrictCSVPayoutReportSummarizer) TryParseUrl(url string, filter func(entry PayoutEntry) bool, resultch chan uint, errch chan error) {
	response, err := http.Get(url)

	if err != nil {
		errch <- fmt.Errorf("HTTP request error: %w", err)
		return
	}

	if response.StatusCode != http.StatusOK {
		errch <- fmt.Errorf("Server returned non-200 status code")
		return
	}

	defer response.Body.Close()

	parser.TryParseStream(response.Body, filter, resultch, errch)
}

func (parser StrictCSVPayoutReportSummarizer) TryParseStream(stream io.Reader, filter func(entry PayoutEntry) bool, resultch chan uint, errch chan error) {
	reader := csv.NewReader(stream)

	header, err := reader.Read()

	if err != nil {
		errch <- fmt.Errorf("CSV read error at line 1: %w", err)
		return
	}

	if header[0] != "date" || header[1] != "name" || header[2] != "amount" {
		errch <- fmt.Errorf("CSV header error")
		return
	}

	var total uint
	nth_line := 1

	for {
		nth_line += 1
		fields, err := reader.Read()

		if err != nil && err.Error() == "EOF" {
			break
		}

		if err != nil {
			errch <- fmt.Errorf("CSV read error at line %d: %w", nth_line, err)
			return
		}

		date, err := time.Parse("2006-01-02", fields[0])

		if err != nil {
			errch <- fmt.Errorf("CSV parse error, invalid date at line %d: %w", nth_line, err)
			return
		}

		if fields[1] == "" {
			errch <- fmt.Errorf("CSV parse error, empty name at line %d", nth_line)
			return
		}

		amount, err := strconv.ParseUint(fields[2], 10, 64)

		if err != nil {
			errch <- fmt.Errorf("CSV parse error, invalid amount at line %d: %w", nth_line, err)
			return
		}

		entry := PayoutEntry{
			date,
			fields[1],
			amount,
		}

		if filter(entry) {
			total += uint(entry.Amount)
		}
	}

	resultch <- total
}

func main() {
	flag_target_date := flag.String("date", time.Now().Format("2006-01-02"), "YYYY-MM-DD. Default: current date.")
	flag_report := flag.String("report", "", "Integer in the range [1, 100]. Overrides -range.")
	flag_range := flag.String("range", "1,100", "Comma-separated pair of integers, e.g 1,10. Endpoints in range [1, 100] and end >= start.")
	flag_task_limit := flag.Uint("limit", 8, "Maximum number of parallell tasks (i.e downloads.")
	flag_debug := flag.Bool("debug", false, "Enable or disable debugging.")

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

	if *flag_task_limit < 1 || *flag_task_limit > 65536 {
		log.Fatalln("FATAL: Invalid task limit. Expected integer in range [1, 65536].")
	}

	task_limit := *flag_task_limit

	var total uint

	parser := StrictCSVPayoutReportSummarizer{}
	num_jobs := target_range_end - target_range_start + 1

	urlch := make(chan string, num_jobs)
	resultch := make(chan uint, num_jobs)
	errch := make(chan error, num_jobs)

	for n := range num_jobs {
		report_num := n + target_range_start
		urlch <- fmt.Sprintf("https://codetest.stim.se/payouts/%d", report_num)
	}

	for n := range task_limit {
		go func() {
			for {
				url, ok := <-urlch

				if !ok {
					return
				}

				if *flag_debug {
					log.Printf("Task runner #%d to run \"%s\".", n, url)
				}

				parser.TryParseUrl(
					url,
					func(entry PayoutEntry) bool {
						return entry.Date == target_date
					},
					resultch,
					errch,
				)
			}
		}()
	}

	if *flag_debug {
		log.Printf("Spawned %d task runner(s).", task_limit)
	}

	for range num_jobs {
		select {
		case partial := <-resultch:
			total += partial
		case err := <-errch:
			log.Fatalf("FATAL: %s", err)
		}
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
