# goe-report

`goe-report` is a command-line interface (CLI) tool designed to interact with the **go-e Wallbox Cloud API v3**. It allows you to quickly fetch the current status of your charging station and generate comprehensive, formatted charging history reports (in terminal or as PDF files) based on specific months and RFID chips.

## Features

- **Status Check**: Real-time insights into your go-e charger (vehicle state, power, temperature, phases, etc.).
- **Charging Reports**: Aggregate charging sessions per month.
- **Filtering**: Optionally filter reports by specific RFID chip IDs or names.
- **Cost Calculation**: Configure your electricity price per kWh to automatically calculate the cost of each session and the total cost.
- **License Plate Mapping**: Associate a vehicle license plate with the generated reports for billing or reimbursement purposes.
- **PDF Export**: Generate aesthetic, ready-to-print PDF reports.
- **Colorized Interface**: Clean, color-coded terminal output for success logs and error handling.

## Installation

### Prerequisites

- [Go](https://go.dev/dl/) v1.18 or higher.
- A **go-e Wallbox** with Cloud API enabled.
- Your Wallbox **Serial Number** and **Cloud API Token**.

### Build from Source

Clone the repository and build the binary using the provided `Makefile`:

```bash
git clone https://github.com/yourusername/goe-report.git
cd goe-report
make build
```

The compiled binary will be placed in the `bin/` directory.

## Configuration

Before you can pull data from your wallbox, you must configure your API credentials. The CLI uses `viper` to store these settings locally in your home directory (`~/.goe-report.yaml` or `.json`), or you can override them using environment variables prefixed with `GOEREPORT_` (e.g. `GOEREPORT_TOKEN`).

### 1. Set API Token & Serial Number

```bash
./bin/goe-report token set YOUR_API_TOKEN
./bin/goe-report serial set 123456
```

### 2. Set Billing Information (Optional)

You can configure a vehicle license plate and the current electricity price (in Euros per kWh) to include cost calculations in your reports.

```bash
./bin/goe-report licenseplate set "B-EV 1234"
./bin/goe-report kwhprice set 0.38
```

To view your current configurations, you can use the `get` subcommands:

```bash
./bin/goe-report token get
./bin/goe-report serial get
./bin/goe-report licenseplate get
./bin/goe-report kwhprice get
```

## Usage

### Check Wallbox Status

To get a real-time overview of your wallbox, including temperatures, charging statistics, and phase details:

```bash
./bin/goe-report status
```

### Generate a Charging Report

To generate a charging history report, you must provide a target month using the `--month` flag in `MM-YYYY` format.

**Terminal Output:**

```bash
./bin/goe-report report --month=02-2026
```

**Terminal Output with RFID Filter:**
If multiple users or cars use the same wallbox, you can filter the output by providing a comma-separated list of RFID tag IDs or Names.

```bash
./bin/goe-report report --month=02-2026 --chipIds=1,ChipName
```

**PDF Export:**
If you need to distribute or print the report, append the `--pdf` flag. A PDF file containing all charging sessions, units, and billing costs will be generated in your current directory.

```bash
./bin/goe-report report --month=02-2026 --pdf
```

## Development

- `make build` - Compiles the `goe-report` binary into the `bin/` directory.
- `make run` - Compiles and immediately executes the binary.
- `make clean` - Removes the `bin/` directory and any cached builds.

## Libraries Used

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework.
- [spf13/viper](https://github.com/spf13/viper) - Configuration management.
- [fatih/color](https://github.com/fatih/color) - Terminal colored output.
- [jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf) - PDF document generation.

## License

This project is licensed under the [MIT License](LICENSE).
