package main

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/sentinelabs/vertex-synapse-grafana/pkg/plugin"
)

func main() {
	// Start listening to requests sent from Grafana.
	if err := datasource.Manage("vertex-synapse-datasource", plugin.NewDatasource, datasource.ManageOpts{}); err != nil {
		log.DefaultLogger.Error(err.Error())
		os.Exit(1)
	}
}
