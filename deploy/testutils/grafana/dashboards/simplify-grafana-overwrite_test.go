//go:build dashboard && simplifydashboard

package dashboard

import (
	"path/filepath"
	"testing"
)

// TestOverwriteDashboards simplifies and overwrites Grafana dashboards in deploy folder variants
func TestOverwriteDashboards(t *testing.T) {
	// get all json's in various generation deploly folders
	files, err := filepath.Glob("../../../*/grafana/dashboards/*.json")

	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		t.Logf("simplifying/overwriting dashboard: %s", file)

		sourcePath := file
		_ = SimplifyGrafana(sourcePath, true)
	}
}
