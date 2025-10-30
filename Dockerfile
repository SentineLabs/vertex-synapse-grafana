# Frontend build stage
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source files needed for webpack build
COPY tsconfig.json webpack.config.js .eslintrc.js ./
COPY plugin.json README.md ./
COPY src/ ./src/

# Build frontend
RUN npm run build

# Backend build stage
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go files
COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/

# Build backend for multiple platforms
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_linux_amd64 ./pkg
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_linux_arm64 ./pkg
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_darwin_amd64 ./pkg
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_darwin_arm64 ./pkg
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_windows_amd64.exe ./pkg

# Final stage
FROM grafana/grafana:latest

# Copy plugin metadata to root
COPY plugin.json /var/lib/grafana/plugins/vertex-synapse-datasource/
COPY README.md /var/lib/grafana/plugins/vertex-synapse-datasource/

# Copy built frontend from frontend-builder
COPY --from=frontend-builder /app/dist/module.js* /var/lib/grafana/plugins/vertex-synapse-datasource/
COPY --from=frontend-builder /app/dist/img/ /var/lib/grafana/plugins/vertex-synapse-datasource/img/
COPY --from=frontend-builder /app/dist/dashboards/ /var/lib/grafana/plugins/vertex-synapse-datasource/dashboards/

# Copy all built backend binaries from backend-builder to root
COPY --from=backend-builder /app/dist/* /var/lib/grafana/plugins/vertex-synapse-datasource/

# Set permissions
USER root
RUN chmod +x /var/lib/grafana/plugins/vertex-synapse-datasource/gpx_vertex-synapse-datasource_*
RUN chown -R grafana:root /var/lib/grafana/plugins/vertex-synapse-datasource

ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=vertex-synapse-datasource
ENV GF_SECURITY_ADMIN_PASSWORD=admin
ENV GF_INSTALL_PLUGINS=""

# Switch to grafana user
USER grafana

EXPOSE 3000
