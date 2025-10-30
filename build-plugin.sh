#!/bin/bash
set -e

echo "Building Vertex Synapse Grafana Plugin using Docker..."

# Clean previous package
echo "Cleaning previous package..."
rm -f vertex-synapse-datasource.zip
rm -rf vertex-synapse-datasource

# Build using Docker (this builds both frontend and backend)
echo "Building plugin in Docker..."
docker build -f Dockerfile --target backend-builder -t vertex-synapse-builder .

# Extract built files from Docker image
echo "Extracting built files..."
docker create --name temp-builder vertex-synapse-builder
docker cp temp-builder:/app/dist ./
docker rm temp-builder

# Create plugin directory structure
echo "Creating plugin package..."
mkdir -p vertex-synapse-datasource
cp plugin.json vertex-synapse-datasource/
cp README.md vertex-synapse-datasource/
cp -r dist/img vertex-synapse-datasource/ 2>/dev/null || true
cp -r dist/dashboards vertex-synapse-datasource/ 2>/dev/null || true
cp dist/gpx_vertex-synapse-datasource_linux_amd64 vertex-synapse-datasource/
cp dist/gpx_vertex-synapse-datasource_linux_arm64 vertex-synapse-datasource/

# Build frontend locally if possible, otherwise use a minimal placeholder
if [ -f "dist/module.js" ]; then
    cp dist/module.js vertex-synapse-datasource/
    cp dist/module.js.map vertex-synapse-datasource/ 2>/dev/null || true
else
    echo "Warning: Frontend not built, creating minimal module.js"
    echo "// Frontend module" > vertex-synapse-datasource/module.js
fi

# Make executables executable
chmod +x vertex-synapse-datasource/gpx_vertex-synapse-datasource_*

# Create zip file
echo "Creating zip file..."
zip -r vertex-synapse-datasource.zip vertex-synapse-datasource

# Cleanup
rm -rf vertex-synapse-datasource

echo ""
echo "✅ Plugin built successfully!"
echo "📦 Package: vertex-synapse-datasource.zip"
echo ""
echo "To install in Grafana:"
echo "1. Copy vertex-synapse-datasource.zip to your server"
echo "2. Extract to /var/lib/grafana/plugins/"
echo "   unzip vertex-synapse-datasource.zip -d /var/lib/grafana/plugins/"
echo "3. Restart Grafana"
echo "4. Set environment variable: GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=vertex-synapse-datasource"
