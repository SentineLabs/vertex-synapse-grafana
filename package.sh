#!/bin/bash
set -e

echo "Packaging Vertex Synapse Grafana Plugin from Docker image..."

# Clean previous package and tmp directory
echo "Cleaning previous package..."
rm -f vertex-synapse-datasource.zip
rm -rf tmp

# Create tmp directory structure
mkdir -p tmp/extracted tmp/package

# Extract plugin files from running Docker container
echo "Extracting plugin files from Grafana container..."
docker cp vertex-synapse-grafana-grafana-1:/var/lib/grafana/plugins/vertex-synapse-datasource tmp/extracted/

# Create clean plugin directory
echo "Creating plugin package..."
mkdir -p tmp/package/vertex-synapse-datasource

# Copy only the necessary files from the extracted directory
cp -r tmp/extracted/vertex-synapse-datasource/module.js* tmp/package/vertex-synapse-datasource/ 2>/dev/null || true
cp -r tmp/extracted/vertex-synapse-datasource/gpx_* tmp/package/vertex-synapse-datasource/ 2>/dev/null || true
cp -r tmp/extracted/vertex-synapse-datasource/img tmp/package/vertex-synapse-datasource/ 2>/dev/null || true

# Copy files from project root (these will override any from the extracted dir)
cp README.md tmp/package/vertex-synapse-datasource/
cp plugin.json tmp/package/vertex-synapse-datasource/
cp LICENSE tmp/package/vertex-synapse-datasource/
cp CHANGELOG.md tmp/package/vertex-synapse-datasource/

# Copy screenshots if they exist
if [ -d "src/img/screenshots" ]; then
  mkdir -p tmp/package/vertex-synapse-datasource/img/screenshots
  cp src/img/screenshots/*.png tmp/package/vertex-synapse-datasource/img/screenshots/ 2>/dev/null || true
fi

# Copy dashboards from source
if [ -d "src/dashboards" ]; then
  mkdir -p tmp/package/vertex-synapse-datasource/dashboards
  cp src/dashboards/*.json tmp/package/vertex-synapse-datasource/dashboards/ 2>/dev/null || true
fi

cp -r tmp/extracted/vertex-synapse-datasource/dashboards tmp/package/vertex-synapse-datasource/ 2>/dev/null || true
cp tmp/extracted/vertex-synapse-datasource/module.js.map tmp/package/vertex-synapse-datasource/ 2>/dev/null || true
cp tmp/extracted/vertex-synapse-datasource/gpx_vertex-synapse-datasource_linux_amd64 tmp/package/vertex-synapse-datasource/
cp tmp/extracted/vertex-synapse-datasource/gpx_vertex-synapse-datasource_linux_arm64 tmp/package/vertex-synapse-datasource/

# Make executables executable
chmod +x tmp/package/vertex-synapse-datasource/gpx_vertex-synapse-datasource_*

# Create zip file
echo "Creating zip file..."
cd tmp/package
zip -r ../../vertex-synapse-datasource.zip vertex-synapse-datasource
cd ../..

# Cleanup
echo "Cleaning up temporary files..."
rm -rf tmp/extracted tmp/package

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
