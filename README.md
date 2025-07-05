# Go Money

![build workflow](https://github.com/ft-t/go-money/actions/workflows/general.yaml/badge.svg?branch=master)
[![codecov](https://codecov.io/gh/ft-t/go-money/graph/badge.svg?token=pas79tP0Dr)](https://codecov.io/gh/ft-t/go-money)
[![go-report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ft-t/go-money)](https://pkg.go.dev/github.com/ft-t/go-money?tab=doc)

**Go Money** is an open-source personal finance manager written in Go, built for speed, flexibility, and advanced customization.

Unlike most finance apps, Go Money is designed for more advanced users who want full control over their transaction data and storage.

It enables customizations through Lua scripting and external reporting with Grafana.

## Key Features

- Support for multi currency transactions
- Custom Lua hooks to process transactions
- Grafana-based reporting (bring your own dashboards)
- Import data from other finance apps (Firefly for now)
- Scriptable and developer-friendly architecture
- High test coverage and stable api
- Multiple client libraries provided via [ConnectRPC](https://buf.build/xskydev/go-money-pb/sdks/main:protobuf) 

## Installation
Go Money is available as a Docker image, making it easy to deploy and run on any system that supports Docker.
For detailed installation instructions, please refer to the [Installation guide](https://github.com/ft-t/go-money/wiki/Installation).

## UI
GO Money provides a simple web UI for managing transactions, accounts, and other financial data.

## API
Go Money provides multi-protocol API (gRPC, JSON-RPC) for more details and documentation, please refer to the [API documentation](https://github.com/ft-t/go-money/wiki/Api)

## Reporting
Go Money does not come with built-in reports. Instead, it allows you to use Grafana to create custom dashboards and reports based on your transaction data.

[Grafana guide](https://github.com/ft-t/go-money/wiki/Grafana)
[Database schema](https://github.com/ft-t/go-money/wiki/Database-structure-and-entities-rules)

[Query examples](https://github.com/ft-t/go-money/tree/master/docs/reporting/queries)

[//]: # ([Grafana dashboards]&#40;https://github.com/ft-t/go-money/tree/master/docs/reporting/dashboards&#41;.)

## Scripting 
Go Money allows you to write Lua scripts to process transactions. This makes it highly flexible and adaptable to your specific needs.

[Lua scripting guide](https://github.com/ft-t/go-money/wiki/Lua)

[Lua scripts examples](https://github.com/ft-t/go-money/tree/master/docs/lua)

## Documentation

Full documentation and examples are available in the [Wiki](https://github.com/ft-t/go-money/wiki)
