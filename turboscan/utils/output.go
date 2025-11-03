package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"turboscan/scanner"
)

// PrintResults prints scan results in a tabular form ordered by status code and URL.
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

// SaveResults stores scan results in a plain text format.
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

// SaveJSON writes scan results to a JSON file.
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
