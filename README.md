# Vertex Synapse Datasource for Grafana

Query [Vertex Synapse](https://vertex.link/synapse) hypergraph data using Storm queries in Grafana.

## Features

- Execute Storm queries via `/api/v1/storm` (streaming) and `/api/v1/storm/call` (single result) endpoints
- Secure API key authentication (backend-only) - works with both Cortex and Optic HTTP API endpoints
- Automatic Grafana time range injection as Storm variables
- Support for essential Storm data types: nodes, objects, lists, primitives

## Installation

### Manual Installation

1. Download the latest `vertex-synapse-datasource.zip` from releases or build it:
   ```bash
   ./package.sh
   ```

2. Copy the zip file to your Grafana server and extract it:
   ```bash
   unzip vertex-synapse-datasource.zip -d /var/lib/grafana/plugins/
   ```

3. Configure Grafana to allow unsigned plugins by setting the environment variable:
   ```bash
   GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=vertex-synapse-datasource
   ```

4. Restart Grafana:
   ```bash
   # For systemd
   sudo systemctl restart grafana-server
   
   # For init.d
   sudo service grafana-server restart
   ```

### Docker Compose (Development)

```bash
# Start services
docker compose up -d

# Create Grafana API user and get API key
docker compose run --rm create-apikey

# Access Grafana at http://localhost:3000 (admin/admin)
# Configure datasource with the API key from above
```

## Configuration

1. In Grafana, go to Configuration > Data Sources
2. Click "Add data source"
3. Search for "Vertex Synapse" and select it
4. Configure the following settings:
   - **Name**: A name for your data source
   - **URL**: The URL of your Vertex Synapse API (e.g., `http://synapse:4443`)
   - **API Key**: Your Synapse API key (recommended: use Grafana secrets)

## Usage

### Query Editor

Use the query editor to write and execute Storm queries:

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

### Time Variables

Automatically injected from Grafana time picker:

- `$timeRange` - Tuple for `@=` queries: `['start', 'end']`
- `$timeFrom`, `$timeTo` - ISO 8601 strings
- `$dateFrom`, `$dateTo` - Date strings (YYYY-MM-DD)
- `$timeFromMs`, `$timeToMs` - Unix milliseconds

## Additional Resources

- [Synapse Documentation](https://synapse.docs.vertex.link/) - Official Synapse documentation
- [Vertex Synapse Grafana Plugin](https://github.com/sentinelabs/vertex-synapse-grafana) - Source code and issue tracking
