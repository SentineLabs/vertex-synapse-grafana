.PHONY: all clean build-frontend build-backend build install dev

all: build

clean:
	rm -rf dist/
	rm -f pkg/plugin/plugin

build-frontend:
	npm ci
	npm run build

build-backend:
	go mod tidy
	CGO_ENABLED=0 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_linux_amd64 ./pkg
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_darwin_amd64 ./pkg
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_darwin_arm64 ./pkg
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -buildvcs=false -o dist/gpx_vertex-synapse-datasource_windows_amd64.exe ./pkg

build: clean build-frontend build-backend

install: build
	cp -r dist/ /var/lib/grafana/plugins/vertex-synapse-datasource/

dev:
	npm run dev
