package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"expressscan/scanner"
	"expressscan/utils"
)

func main() {
	url := flag.String("u", "", "Target URL")
	wordlistPath := flag.String("w", "", "Wordlist path")
	threads := flag.Int("t", 50, "Number of threads")
	timeout := flag.Int("timeout", 10, "Request timeout (seconds)")
	statusCodes := flag.String("mc", "200,301,302,401,403", "Match status codes (comma separated)")
	outputPath := flag.String("o", "", "Output file (plain text)")
	jsonPath := flag.String("json", "", "Optional JSON output file")
	verbose := flag.Bool("v", false, "Verbose mode")
	recursive := flag.Bool("r", false, "Enable recursive scanning")
	maxDepth := flag.Int("depth", 1, "Maximum recursion depth")
	rateLimit := flag.Int("rate", 0, "Maximum requests per second (0 = unlimited)")
	retries := flag.Int("retries", 2, "Maximum retries for transient errors")
	extensions := flag.String("e", "", "Comma separated list of extensions to brute-force")
	disableFilter := flag.Bool("no-filter", false, "Disable smart 404 filtering")

	flag.Parse()

	if *url == "" || *wordlistPath == "" {
		fmt.Println("Usage: expressscan -u <url> -w <wordlist>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	printBanner()

	words, err := utils.LoadWordlist(*wordlistPath)
	if err != nil {
		fmt.Printf("Error loading wordlist: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Loaded %d words\n", len(words))
	fmt.Printf("[+] Target: %s\n", *url)
	fmt.Printf("[+] Threads: %d\n", *threads)

	cfg := scanner.Config{
		BaseURL:           strings.TrimRight(*url, "/"),
		Threads:           *threads,
		Timeout:           *timeout,
		StatusCodes:       parseStatusCodes(*statusCodes),
		Verbose:           *verbose,
		MaxRetries:        *retries,
		RateLimit:         *rateLimit,
		Recursive:         *recursive,
		MaxDepth:          *maxDepth,
		Extensions:        parseExtensions(*extensions),
		EnableSmartFilter: !*disableFilter,
	}

	sc := scanner.NewScanner(cfg)

	var results []scanner.Result
	if cfg.Recursive {
		results = sc.ScanRecursive(words, cfg.MaxDepth)
	} else if len(cfg.Extensions) > 0 {
		results = sc.ScanWithExtensions(words, cfg.Extensions)
	} else {
		results = sc.Scan(words)
	}

	utils.PrintResults(results)

	if *outputPath != "" {
		if err := utils.SaveResults(results, *outputPath); err != nil {
			fmt.Printf("Error saving results: %v\n", err)
		} else {
			fmt.Printf("[+] Results saved to %s\n", *outputPath)
		}
	}

	if *jsonPath != "" {
		if err := utils.SaveJSON(results, *jsonPath); err != nil {
			fmt.Printf("Error saving JSON: %v\n", err)
		} else {
			fmt.Printf("[+] JSON saved to %s\n", *jsonPath)
		}
	}

	sc.PrintStats()
}

func parseStatusCodes(codes string) []int {
	var result []int
	for _, part := range strings.Split(codes, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		code, err := utils.ParseInt(part)
		if err != nil {
			fmt.Printf("[!] Invalid status code '%s': %v\n", part, err)
			continue
		}
		result = append(result, code)
	}
	return result
}

func parseExtensions(ext string) []string {
	if ext == "" {
		return nil
	}
	pieces := strings.Split(ext, ",")
	out := make([]string, 0, len(pieces))
	for _, p := range pieces {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		out = append(out, p)
	}
	return out
}

func printBanner() {
	fmt.Println(`
████████╗██╗   ██╗██████╗ ██████╗  ██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗
╚══██╔══╝██║   ██║██╔══██╗██╔══██╗██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
   ██║   ██║   ██║██████╔╝██████╔╝██║   ██║███████╗██║     ███████║██╔██╗ ██║
   ██║   ██║   ██║██╔══██╗██╔══██╗██║   ██║╚════██║██║     ██╔══██║██║╚██╗██║
   ██║   ╚██████╔╝██║  ██║██████╔╝╚██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝
                    expressscan by 1fanya | v1.0
`)
}
