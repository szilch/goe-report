![Build Status](https://github.com/szilch/echarge-report/actions/workflows/go.yml/badge.svg)
![Lint Status](https://github.com/szilch/echarge-report/actions/workflows/lint-pr.yml/badge.svg)
![Release Status](https://github.com/szilch/echarge-report/actions/workflows/release.yml/badge.svg)
[![codecov](https://codecov.io/gh/szilch/echarge-report/graph/badge.svg)](https://codecov.io/gh/szilch/echarge-report)

<img src="logo.png" alt="Logo" width="128"/>

# echarge-report

CLI tool for interacting with **wallbox charging stations** — fetch real-time status and generate charging reports (terminal or PDF).

## Features

- **Real-time Status**: Fetch the current state of your wallbox.
- **Detailed Reports**: Generate charging reports for specific months or ranges.
- **PDF Export & Merge**: Export to PDF and attach your electricity contracts.
- **RFID Filtering**: Filter sessions by RFID chip IDs or names.
- **Mail Support**: Send reports directly via email.
- **Smarthome Integration**: Include vehicle mileage from Home Assistant.

**Supported Wallboxes:**
- **go-e Charger** (Cloud API v3 and Local API)
- *More coming soon (OpenWB, easee, etc.)*

## Quick Start

1. **Installation**: Download the [latest binary](https://github.com/szilch/echarge-report/releases) or see the [[Installation]] guide.
2. **Configuration**: 
   ```bash
   echarge-report config-set wallbox.goe.cloud.token YOUR_TOKEN
   echarge-report config-set wallbox.goe.cloud.serial 123456
   ```
   Check the [[Configuration]] wiki for more details and a full `config.yml` example.

## Usage

```bash
# Show status
echarge-report status

# Generate report for previous month
echarge-report report --pdf --attach-pdfs
```

## Documentation

For full documentation, including Docker setup and developer guides, please visit our **[GitHub Wiki](https://github.com/szilch/echarge-report/wiki)**.

- [[Installation]]
- [[Configuration]]
- [[Using Docker]]
- [[Developer Documentation]]

## License

[MIT](LICENSE)
