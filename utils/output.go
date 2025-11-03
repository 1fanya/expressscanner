package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"ultrascan/scanner"
)

// PrintResults renders scan findings in a sorted table for quick triage.
func PrintResults(results []scanner.Result) {
	if len(results) == 0 {
		fmt.Println("[!] No results found")
		return
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].StatusCode == results[j].StatusCode {
			return results[i].URL < results[j].URL
		}
		return results[i].StatusCode < results[j].StatusCode
	})

	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "STATUS\tURL\tSIZE\tTIME\tREDIRECT")
	for _, result := range results {
		size := result.Size
		if size < 0 {
			size = 0
		}
		fmt.Fprintf(
			tw,
			"%d\t%s\t%d\t%s\t%s\n",
			result.StatusCode,
			result.URL,
			size,
			formatDuration(result.Time),
			result.RedirectLocation,
		)
	}
	tw.Flush()
}

// SaveResults writes the findings to a newline-delimited plain text file.
func SaveResults(results []scanner.Result, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, result := range results {
		if _, err := fmt.Fprintf(file, "%d %s %d %s %s\n", result.StatusCode, result.URL, result.Size, formatDuration(result.Time), result.RedirectLocation); err != nil {
			return err
		}
	}

	return nil
}

// SaveJSON serialises the results to prettified JSON for downstream tooling.
func SaveJSON(results []scanner.Result, path string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	return d.Round(time.Millisecond).String()
}
