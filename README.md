# UltraScan

UltraScan is a professional-grade web directory enumerator engineered by 1fanya. Built in Go, it delivers aggressive throughput with meticulous resource management so reconnaissance teams can map attack surfaces significantly faster than legacy scanners. UltraScan balances speed with signal fidelity, reducing noise while revealing the endpoints that matter.

## Why UltraScan
- **Performance headroom** – Highly tuned HTTP transport, wide connection pools, and HTTP/2 support keep pipelines saturated even on large wordlists.
- **Resilient discovery** – Automatic retry logic and optional rate limiting maintain stability on fragile infrastructure without sacrificing coverage.
- **Noise reduction** – Adaptive 404 calibration and extension-aware filtering help distinguish genuine findings from false positives.
- **Depth when you need it** – Recursive exploration with configurable depth follows redirects and newly exposed paths automatically.
- **Operator-friendly output** – Real-time console updates, tabular summaries, and both plaintext and JSON exports fit easily into reporting workflows.

## Installation
1. Install Go 1.20 or later.
2. Clone this repository and change into the project directory.
3. Build the binary:
   ```bash
   go build -o ultrascan
   ```

## Quick start
Kick off an initial reconnaissance run with:
```bash
./ultrascan -u https://example.com -w wordlists/common.txt
```

### Frequently used flags
- `-t` – number of concurrent workers (default 50).
- `-timeout` – HTTP request timeout in seconds (default 10).
- `-mc` – comma-separated list of status codes to report.
- `-o` – save results to a plaintext file.
- `-json` – persist structured output for automation.
- `-v` – verbose logging for troubleshooting failed requests.
- `-retries` – retry budget for transient network failures (default 2).
- `-rate` – throttle requests per second (0 disables throttling).
- `-r` / `-depth` – enable recursive scanning and define maximum depth.
- `-e` – comma-separated extensions appended to each word.
- `-no-filter` – disable smart 404 detection.

## Advanced playbooks
High-concurrency scan with retries and saved output:
```bash
./ultrascan -u https://example.com -w lists/common.txt -t 120 -retries 3 -o hits.txt
```

Recursive exploration with extension brute forcing and JSON export:
```bash
./ultrascan -u https://example.com -w lists/common.txt -r -depth 3 -e php,html,txt -json results.json
```

Measured probing with rate limiting and verbose feedback:
```bash
./ultrascan -u https://example.com -w lists/large.txt -rate 75 -v
```

## Output model
UltraScan streams every confirmed hit as it happens and concludes with an aligned summary table. Each row details the status code, URL, response size, request duration, and redirect target when applicable. File exports mirror the same metadata, ensuring parity between live operations and stored artifacts.

## License
UltraScan is distributed under the MIT License. See [LICENSE](LICENSE) for the full text.
