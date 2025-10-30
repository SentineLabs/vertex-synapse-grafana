#!/bin/bash
set -e

echo "Packaging Vertex Synapse Grafana Plugin from Docker image..."

# Clean previous package
echo "Cleaning previous package..."
rm -f vertex-synapse-datasource.zip
rm -rf vertex-synapse-datasource
rm -rf temp-plugin-files

# Extract plugin files from running Docker container
echo "Extracting plugin files from Grafana container..."
docker cp vertex-synapse-grafana-grafana-1:/var/lib/grafana/plugins/vertex-synapse-datasource ./temp-plugin-files

# Create clean plugin directory
echo "Creating plugin package..."
mkdir -p vertex-synapse-datasource
cp temp-plugin-files/plugin.json vertex-synapse-datasource/
cp temp-plugin-files/README.md vertex-synapse-datasource/
cp -r temp-plugin-files/img vertex-synapse-datasource/
cp temp-plugin-files/module.js vertex-synapse-datasource/
cp temp-plugin-files/module.js.map vertex-synapse-datasource/ 2>/dev/null || true
cp temp-plugin-files/gpx_vertex-synapse-datasource_linux_amd64 vertex-synapse-datasource/
cp temp-plugin-files/gpx_vertex-synapse-datasource_linux_arm64 vertex-synapse-datasource/

# Make executables executable
chmod +x vertex-synapse-datasource/gpx_vertex-synapse-datasource_*

# Create zip file
echo "Creating zip file..."
zip -r vertex-synapse-datasource.zip vertex-synapse-datasource

# Cleanup
rm -rf vertex-synapse-datasource temp-plugin-files

echo ""
echo "✅ Plugin packaged successfully!"
echo "📦 Package: vertex-synapse-datasource.zip"
echo ""
echo "To install in Grafana:"
echo "1. Copy vertex-synapse-datasource.zip to your server"
echo "2. Extract to /var/lib/grafana/plugins/"
echo "   unzip vertex-synapse-datasource.zip -d /var/lib/grafana/plugins/"
echo "3. Restart Grafana"
echo "4. Set environment variable: GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=vertex-synapse-datasource"
