# Gov

**A simple prometheus metrics and pprof profile viewer.**

![Screenshot](http://joel-github-static.s3.amazonaws.com/gov/screenshot.png)

## Installation

Use Go to install the binary:

```bash
go install github.com/256dpi/gov
```

## Usage

Run gov with the URL of the program to collect metrics and profiles from.

```gov
gov http://localhost:1234
```

Prometheus metrics are collected from the `/metrics` endpoint while pprof
profiles are collected from the `/debug/pprof/{profile,allocs,heap,block,mutex}`
endpoints.
