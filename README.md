![Build Status](https://github.com/szilch/echarge-report/actions/workflows/go.yml/badge.svg)
![Lint Status](https://github.com/szilch/echarge-report/actions/workflows/lint-pr.yml/badge.svg)
![Release Status](https://github.com/szilch/echarge-report/actions/workflows/release.yml/badge.svg)
[![codecov](https://codecov.io/gh/szilch/echarge-report/graph/badge.svg)](https://codecov.io/gh/szilch/echarge-report)

![GitHub License](https://img.shields.io/github/license/szilch/echarge-report)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/szilch/echarge-report)

<img src="pkg/formatter/logo.png" alt="Logo" width="128"/>

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
- _More coming soon (OpenWB, easee, etc.)_

## Quick Start

1. **Installation**: Download the [latest binary](https://github.com/szilch/echarge-report/releases) or see the [Installation](https://github.com/szilch/echarge-report/wiki/Installation) guide.
2. **Configuration**:
   ```bash
   echarge-report config-set wallbox.goe.cloud.token YOUR_TOKEN
   echarge-report config-set wallbox.goe.cloud.serial 123456
   echarge-report config-set wallbox.goe.chipIds 12345,67890
   echarge-report config-set licenseplate BIT-SZ-58E
   echarge-report config-set driver "Max Mustermann"
   ```
   Check the [Configuration](https://github.com/szilch/echarge-report/wiki/Configuration) wiki for more details and a full [config.yml](https://github.com/szilch/echarge-report/wiki/Configuration#example-configyml) example.
3. **Generate Report**:

## Usage

```bash
# Show version
echarge-report version

# Show status of wallbox
echarge-report status

# Generate report for previous month
echarge-report report

# Generate report for a specific month as pdf
echarge-report report --month 03-2026 --pdf
```

## Documentation

For full documentation, including Docker setup and developer guides, please visit our **[GitHub Wiki](https://github.com/szilch/echarge-report/wiki)**.

- [User Documentation](https://github.com/szilch/echarge-report/wiki/User-Documentation)
- [Developer Documentation](https://github.com/szilch/echarge-report/wiki/Developer-Documentation)

## License

[MIT](LICENSE)
