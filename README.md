![Build Status](https://github.com/szilch/goe-report/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/szilch/goe-report/graph/badge.svg)](https://codecov.io/gh/szilch/goe-report)

# goe-report

CLI tool for interacting with **wallbox charging stations** — fetch real-time status and generate charging reports (terminal or PDF).

## 1. Overview

`goe-report` is a flexible tool to easily gather insights and generate reports from your wallbox charger.

**Supported Wallboxes:**
- **go-e Charger** (Cloud API v3 and Local API)
- More wallbox types coming soon! The architecture supports easy extension via adapters.

**Feature Overview:**

- **Real-time Status**: Fetch the current state of your go-e Charger.
- **Detailed Reports**: Generate charging reports for specific months or date ranges.
- **PDF Export**: Easily export tabular charging reports into a neat PDF document.
- **Merge PDFs**: Automatically append existing PDFs (e.g. your electricity contract `.pdf` files) to the generated report.
- **RFID Filtering**: Filter your charging sessions by specific RFID chip IDs or names.
- **Mail Support**: Send generated PDFs directly via email.
- **Home Assistant Integration**: Fetch the current mileage of your EV from Home Assistant and include it in your reports.

**Downloads & Packages:**

- [GitHub Releases (Linux, Windows, macOS binaries)](https://github.com/szilch/goe-report/releases)
- [GitHub Container Registry (Docker Packages)](https://github.com/szilch/goe-report/pkgs/container/goe-report)

---

## 2. For Users

### Configuration

Settings are stored in `~/.goe-report/.goereportrc` or can simply be set using the CLI `config-set` commands. They can also be set via environment variables prefixed with `GOEREPORT_` (e.g. `GOEREPORT_WALLBOX_TOKEN`).

**Important:** You must configure either the **Cloud API** (`wallbox_token` and `wallbox_serial`) OR the **Local API** (`wallbox_localApiUrl`). You do not need both.

#### General Settings

| Parameter / Key  | Environment Variable          | Requirement | Description                                                            |
| ---------------- | ----------------------------- | ----------- | ---------------------------------------------------------------------- |
| `wallbox_type`   | `GOEREPORT_WALLBOX_TYPE`      | Optional    | Wallbox type (e.g. `goe`). Defaults to `goe`. More types coming soon.  |
| `licenseplate`   | `GOEREPORT_LICENSEPLATE`      | Optional    | License plate to show on the report                                    |
| `kwhprice`       | `GOEREPORT_KWHPRICE`          | Optional    | Price per kWh (e.g., `0.38`)                                           |

#### Wallbox Connection Settings

| Parameter / Key      | Environment Variable            | Requirement          | Description                                                     |
| -------------------- | ------------------------------- | -------------------- | --------------------------------------------------------------- |
| `wallbox_token`      | `GOEREPORT_WALLBOX_TOKEN`       | **Required (Cloud)** | Your Wallbox Cloud API Token (e.g. go-e Cloud Token)            |
| `wallbox_serial`     | `GOEREPORT_WALLBOX_SERIAL`      | **Required (Cloud)** | Your wallbox serial number                                      |
| `wallbox_localApiUrl`| `GOEREPORT_WALLBOX_LOCALAPIURL` | **Required (Local)** | The URL to your local Wallbox API (e.g., `http://192.168.1.50`) |
| `wallbox_chipIds`    | `GOEREPORT_WALLBOX_CHIPIDS`     | Optional             | Comma-separated list of RFID chips to filter (e.g., `1,MyChip`) |

#### Home Assistant Settings

| Parameter / Key      | Environment Variable            | Requirement          | Description                                                     |
| -------------------- | ------------------------------- | -------------------- | --------------------------------------------------------------- |
| `ha_api`             | `GOEREPORT_HA_API`              | Optional             | Home Assistant URL (e.g., `http://homeassistant.local:8123`)    |
| `ha_token`           | `GOEREPORT_HA_TOKEN`            | Optional             | Home Assistant Long-Lived Access Token                          |
| `ha_milage_sensorid` | `GOEREPORT_HA_MILAGE_SENSORID`  | Optional             | HA Sensor ID for mileage (e.g., `sensor.car_mileage`)           |

#### Mail Settings

| Parameter / Key      | Environment Variable            | Requirement          | Description                                                     |
| -------------------- | ------------------------------- | -------------------- | --------------------------------------------------------------- |
| `mail_host`          | `GOEREPORT_MAIL_HOST`           | Optional             | SMTP Mail Host                                                  |
| `mail_port`          | `GOEREPORT_MAIL_PORT`           | Optional             | SMTP Mail Port (e.g., `587`)                                    |
| `mail_username`      | `GOEREPORT_MAIL_USERNAME`       | Optional             | SMTP Username                                                   |
| `mail_password`      | `GOEREPORT_MAIL_PASSWORD`       | Optional             | SMTP Password                                                   |
| `mail_from`          | `GOEREPORT_MAIL_FROM`           | Optional             | Sender Email                                                    |
| `mail_to`            | `GOEREPORT_MAIL_TO`             | Optional             | Comma-separated recipient emails                                |

### Command Line Interface (CLI)

```bash
# Configuration setup commands
./bin/goe-report config-set wallbox_token YOUR_API_TOKEN
./bin/goe-report config-set wallbox_serial 123456
./bin/goe-report config-list

# Show current wallbox status
./bin/goe-report status

# Charging report for the previous month (terminal output)
./bin/goe-report report

# Charging report for a specific month (terminal output)
./bin/goe-report report --month=02-2026

# Charging report for a date range (multiple months)
./bin/goe-report report --from-month=01-2026 --to-month=03-2026

# Filter by RFID chip ID or name
./bin/goe-report report --month=02-2026 --chipIds=1,MyChip

# Export as PDF
./bin/goe-report report --month=02-2026 --pdf

# Export as PDF and append all existing PDFs found in ~/.goe-report/
./bin/goe-report report --month=02-2026 --pdf --attach-pdfs

# Export as PDF and send it via email (requires mail config)
./bin/goe-report report --month=02-2026 --pdf --send-mail
```

> **Note:** Only RFID tags configured directly on the wallbox can be used for filtering.

### Setup via Docker Compose

You can run `goe-report` periodically as a cron job inside a lightweight Docker container. There are two ways to configure the container: using environment variables in the `docker-compose.yml` file, or by providing your existing `.goereportrc` file via a volume mount.

#### Option A: Using Environment Variables (Recommended for pure Docker)

Configure everything directly within your `docker-compose.yml`:

```yaml
version: "3.8"

services:
  goe-report-cron:
    image: ghcr.io/szilch/goe-report:latest
    container_name: goe-report-cron
    restart: unless-stopped
    volumes:
      # Mount a local 'data' directory to attach existing PDFs
      - ./data:/home/goe-report/.goe-report
    environment:
      # CRON_EXPRESSION syntax: "min hour day month weekday"
      # Default "0 12 1 * *" generates the report on the 1st of every month at 12:00
      - CRON_EXPRESSION=0 12 1 * *

      # This is the CLI command executed periodically
      - CRON_COMMAND=/app/goe-report report --pdf --attach-pdfs --send-mail

      # Define your configuration variables here:
      - GOEREPORT_WALLBOX_SERIAL=your_serial_number
      - GOEREPORT_WALLBOX_TOKEN=your_cloud_token
```

#### Option B: Using a Configuration File

If you have already configured the tool locally, you can simply reuse your `.goereportrc` file.
Place your `.goereportrc` file inside the `./data` folder and mount it into the container. `goe-report` will automatically read it. You only need to define the CRON variables in your `docker-compose.yml`:

```yaml
version: "3.8"

services:
  goe-report-cron:
    image: ghcr.io/szilch/goe-report:latest
    container_name: goe-report-cron
    restart: unless-stopped
    volumes:
      # The container will read your ./data/.goereportrc file
      - ./data:/home/goe-report/.goe-report
    environment:
      - CRON_EXPRESSION=0 12 1 * *
      - CRON_COMMAND=/app/goe-report report --pdf --attach-pdfs --send-mail
      # Everything else is read from the mounted config file
```

#### Starting the Container

1. Create the `data` directory next to your `docker-compose.yml`: `mkdir data`
2. If using Option B, copy your config file: `cp ~/.goe-report/.goereportrc ./data/`
3. **If using `--attach-pdfs`**: Place all PDF files you want to merge (e.g., your electricity contract) directly into the newly created `./data` directory. The container will automatically pick them up.
4. Start the container in the background:
   ```bash
   docker-compose up -d
   ```
5. Check the logs:
   ```bash
   docker-compose logs -f
   ```

---

## 3. For Developers

### Prerequisites

- [Go](https://go.dev/dl/) v1.18+

### Build & Run

| Command      | Description            |
| ------------ | ---------------------- |
| `make build` | Compile the binary     |
| `make run`   | Compile and run        |
| `make clean` | Remove build artifacts |

```bash
git clone https://github.com/szilch/goe-report.git
cd goe-report
make build   # binary is placed in bin/
```

### Docker Build

If you want to build the Docker image locally instead of pulling it from the registry, you can use the Makefile:

```bash
make docker-build
```

### Libraries Used

- [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
- [spf13/viper](https://github.com/spf13/viper) — Configuration management
- [fatih/color](https://github.com/fatih/color) — Colored terminal output
- [jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) — PDF generation
- [pdfcpu/pdfcpu](https://github.com/pdfcpu/pdfcpu) — PDF merging and manipulation

## License

[MIT](LICENSE)
