# logpipe

A structured log aggregator that tails multiple files and forwards parsed JSON logs to configurable sinks.

---

## Installation

```bash
go install github.com/yourname/logpipe@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/logpipe.git && cd logpipe && go build -o logpipe .
```

---

## Usage

Define your sources and sinks in a config file:

```yaml
# logpipe.yaml
sources:
  - path: /var/log/app/*.log
  - path: /var/log/nginx/access.log

sinks:
  - type: stdout
  - type: http
    url: https://logs.example.com/ingest
    headers:
      Authorization: Bearer $LOG_TOKEN
```

Then run:

```bash
logpipe --config logpipe.yaml
```

logpipe will tail all matching files, parse each line as JSON, and forward structured log entries to every configured sink. Non-JSON lines are skipped by default (use `--raw` to forward them as plain strings).

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `logpipe.yaml` | Path to config file |
| `--raw` | `false` | Forward non-JSON lines as raw strings |
| `--poll` | `false` | Use polling instead of inotify |

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss significant changes.

---

## License

[MIT](LICENSE)