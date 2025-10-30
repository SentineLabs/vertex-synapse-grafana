# Vertex Synapse Datasource for Grafana

Query [Vertex Synapse](https://synapse.docs.vertex.link/) hypergraph data using Storm queries in Grafana.

## Features

- Execute Storm queries via `/api/v1/storm` (streaming) and `/api/v1/storm/call` (single result) endpoints
- Secure API key authentication (backend-only) - works with both Cortex and Optic HTTP API endpoints
- Automatic Grafana time range injection as Storm variables
- Support for essential Storm data types: nodes, objects, lists, primitives

## Quick Start

```bash
# Start Cortex and Grafana
docker compose up -d

# Load sample data into Cortex
docker compose run --rm load-data

# Create Grafana API user and get API key
docker compose run --rm create-apikey

# Access Grafana at http://localhost:3000 (admin/admin)
# Configure datasource with the API key from above
```

### Interactive Storm Shell

```bash
# Open interactive Storm shell
docker compose run --rm storm

# In the Storm shell, you can run queries:
storm> inet:fqdn
storm> inet:ipv4 +#malware
storm> $lib.view.list()
```

## Query Examples

```storm
# Basic queries
inet:fqdn                           # List all domains
inet:ipv4 +#malware                 # Find malicious IPs

# Time filtering (uses Grafana time picker)
inet:flow +:time@=$timeRange        # Flows in selected time range
inet:dns:a +:seen@=($timeFrom, $timeTo)  # DNS records seen in range

# Storm Call API (toggle "Use Call API")
return($lib.view.list())            # Return view list
return(({'key': 'value'}))          # Return object
```

## Time Variables

Automatically injected from Grafana time picker:

- `$timeRange` - Tuple for `@=` queries: `['start', 'end']`
- `$timeFrom`, `$timeTo` - ISO 8601 strings
- `$dateFrom`, `$dateTo` - Date strings (YYYY-MM-DD)
- `$timeFromMs`, `$timeToMs` - Unix milliseconds

## Development

```bash
# Frontend
npm install
npm run build

# Backend  
go build -o dist/gpx_vertex-synapse-datasource_$(go env GOOS)_$(go env GOARCH) ./pkg
```
