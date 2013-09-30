package gangliamr

import (
	"testing"

	"github.com/daaku/go.ganglia/gmon"
	"github.com/daaku/go.ganglia/gmondtest"
	"github.com/daaku/go.metrics"
)

func TestHistogramSimple(t *testing.T) {
	t.Parallel()
	h := gmondtest.NewHarness(t)
	defer h.Stop()

	const name = "histogram_simple_metric"
	var hg metrics.Histogram
	hg = &Histogram{
		Name: name,
	}

	registry := testRegistry(h.Client)
	registry.Register(hg)

	const v1 = 43
	hg.Update(v1)
	if hg.Count() != 1 {
		t.Fatalf("was expecting 1 got %d", hg.Count())
	}

	registry.write()
	h.ContainsMetric(&gmon.Metric{
		Name:  name + ".count",
		Value: "1",
		Unit:  "count",
		Slope: "both",
	})

	const v2 = 42
	hg.Update(v2)
	if hg.Count() != 2 {
		t.Fatalf("was expecting 2 got %d", hg.Count())
	}

	registry.write()
	h.ContainsMetric(&gmon.Metric{
		Name:  name + ".count",
		Value: "2",
		Unit:  "count",
		Slope: "both",
	})

	if hg.Max() != v1 {
		t.Fatalf("was expecting %d but got %d", v1, hg.Max())
	}

	h.ContainsMetric(&gmon.Metric{
		Name:  name + ".mean",
		Value: "42.5",
		Unit:  "value",
		Slope: "both",
	})
}
