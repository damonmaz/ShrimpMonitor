# Shrimp Monitor

A Linux resource monitoring tool that reads CPU and memory statistics from procfs and serves the latest measurements over a local JSON API.

## Requirements

- Linux
- Go 1.26.8 or newer

The monitor reads CPU data from `/proc/cpuinfo` and `/proc/stat`, and memory data from `/proc/meminfo`.

## Run

From the repository root, start the monitor with:

```sh
go run .
```

The monitor samples CPU and memory every 500 milliseconds. The API listens on `127.0.0.1:8080`, so it is available locally and is not exposed on other network interfaces.

## API

Both endpoints accept `GET` requests and return JSON:

| Endpoint | Description |
| --- | --- |
| `http://localhost:8080/api/cpu` | CPU model, physical core and logical thread counts, aggregate utilization, and utilization for each logical CPU. |
| `http://localhost:8080/api/memory` | Total, used, and available memory in megabytes. |

Example requests:

```sh
curl http://localhost:8080/api/cpu
curl http://localhost:8080/api/memory
```

CPU utilization values are percentages. The CPU response includes `name`, `cores`, `threads`, `utilization_percent`, and a `logical_cores` array containing each logical CPU's `label` and `utilization_percent`. The memory response includes `total`, `used`, and `available`, measured in MB. Either response may include an `error` field if a monitor encounters a read or parsing error.

## Development

Run the Go package checks with:

```sh
go test ./...
```