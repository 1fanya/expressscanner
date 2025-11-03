# expressscan

expressscan is a high-speed web directory enumerator written in Go by 1fanya. It focuses on efficient networking, responsive output, and flexible filtering so you can map targets quickly without sacrificing accuracy.

## Key capabilities
- Connection pooling with an optimized HTTP client tuned for parallel requests and HTTP/2 when available.
- Worker pool architecture with optional rate limiting to balance throughput and caution.
- Smart 404 detection that calibrates against target-specific not-found responses to reduce noise.
- Recursive discovery with configurable depth to follow redirects and dive deeper automatically.
- Extension brute forcing and configurable status code matching for precise discovery.
- Rich output options including realtime console updates, tabulated summaries, and JSON/file exports.

## Getting started
1. Install Go 1.20 or newer.
2. Clone this repository and enter the project directory.
3. Build the binary:
   ```bash
   go build -o expressscan
   ```

## Basic usage
Run a simple scan by specifying the target URL and a wordlist:
```bash
./expressscan -u https://example.com -w /path/to/wordlist.txt
```

Important flags:
- `-t` — number of worker threads (default 50).
- `-timeout` — request timeout in seconds (default 10).
- `-mc` — comma-separated status codes to report (default `200,301,302,401,403`).
- `-o` — save results to a plain text file.
- `-json` — export results to JSON.
- `-v` — verbose mode prints errors for failed requests.
- `-r` and `-depth` — enable recursive scanning and choose how deep to follow redirects.
- `-rate` — limit requests per second (0 disables throttling).
- `-retries` — number of retries for transient network errors (default 2).
- `-e` — comma-separated extensions to append to every word (e.g. `php,html,txt`).
- `-no-filter` — disable automatic 404 calibration if you prefer raw responses.

## Example workflows
Scan a target with 100 workers, retry on transient errors, and store results:
```bash
./expressscan -u https://example.com -w lists/common.txt -t 100 -retries 3 -o hits.txt
```

Enumerate recursively up to depth 3, testing common extensions and exporting JSON:
```bash
./expressscan -u https://example.com -w lists/common.txt -r -depth 3 -e php,html,txt -json results.json
```

Throttle requests to 75 per second while keeping verbose output:
```bash
./expressscan -u https://example.com -w lists/big.txt -rate 75 -v
```

## Output
expressscan prints findings in real time and summarizes them in a table after the scan. Each result includes the HTTP status, URL, response size, request duration, and redirect target (if present). When saving to files, the same information is persisted for later analysis.

## License
This project is released under the MIT License. See [LICENSE](LICENSE) for details.
