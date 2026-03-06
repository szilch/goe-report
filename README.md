![Build Status](https://github.com/szilch/goe-report/actions/workflows/go.yml/badge.svg)

# goe-report

CLI tool for interacting with the **go-e Wallbox Cloud API v3** — fetch real-time status and generate charging reports (terminal or PDF).

## Prerequisites

- [Go](https://go.dev/dl/) v1.18+
- go-e Wallbox with Cloud API enabled (serial number + API token)

## Build

```bash
git clone https://github.com/yourusername/goe-report.git
cd goe-report
make build   # binary is placed in bin/
```

## Configuration

Settings are stored in `~/.goe-report/.goereportrc`. They can also be set via environment variables prefixed with `GOEREPORT_` (e.g. `GOEREPORT_GOE_TOKEN`).

```bash
# Set a value
./bin/goe-report config-set goe_token      YOUR_API_TOKEN
./bin/goe-report config-set goe_serial     123456
./bin/goe-report config-set licenseplate  "B-EV 1234"
./bin/goe-report config-set kwhprice      0.38

# Home Assistant (optional — shows mileage in report)
./bin/goe-report config-set ha_api               https://homeassistant.local:8123
./bin/goe-report config-set ha_token             YOUR_HA_TOKEN
./bin/goe-report config-set ha_milage_sensorid   sensor.car_mileage

# Mail (optional — for sending reports)
./bin/goe-report config-set mail_host            smtp.example.com
./bin/goe-report config-set mail_port            587
./bin/goe-report config-set mail_username        user@example.com
./bin/goe-report config-set mail_password        secret123
./bin/goe-report config-set mail_from            reports@example.com
./bin/goe-report config-set mail_to              "user1@example.com,user2@example.com"

# Read a single value / show all
./bin/goe-report config-get goe_token
./bin/goe-report config-list
```

## Usage

```bash
# Show current wallbox status
./bin/goe-report status

# Charging report for a given month (terminal output)
./bin/goe-report report --month=02-2026

# Filter by RFID chip ID or name
./bin/goe-report report --month=02-2026 --chipIds=1,MyChip

# Export as PDF
./bin/goe-report report --month=02-2026 --pdf

# Export as PDF and append all existing PDFs found in ~/.goe-report/
./bin/goe-report report --month=02-2026 --pdf --attach-pdfs
```

> **Note:** Only RFID tags configured directly on the wallbox can be used for filtering.

## Development

| Command      | Description            |
| ------------ | ---------------------- |
| `make build` | Compile the binary     |
| `make run`   | Compile and run        |
| `make clean` | Remove build artifacts |

## Libraries Used

- [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
- [spf13/viper](https://github.com/spf13/viper) — Configuration management
- [fatih/color](https://github.com/fatih/color) — Colored terminal output
- [jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) — PDF generation
- [pdfcpu/pdfcpu](https://github.com/pdfcpu/pdfcpu) — PDF merging and manipulation

## License

[MIT](LICENSE)
