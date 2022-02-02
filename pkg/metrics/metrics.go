package metrics

import (
	"time"

	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

var (
	clusterIDLabel = MetricLabel{
		Name: "cluster_id",
		Help: "Cluster ID",
	}
	ClusterID    string
	InstanceFQDN string
)

type GlusterMetric struct {
	Name string
	FN   func(glusterutils.GInterface) error
}

var GlusterMetrics []GlusterMetric

func registerMetric(name string, fn func(glusterutils.GInterface) error) {
	GlusterMetrics = append(GlusterMetrics, GlusterMetric{Name: name, FN: fn})
}

// MetricLabel represents Prometheus Label
type MetricLabel struct {
	Name string
	Help string
}

// Metric represents Prometheus metric
type Metric struct {
	Name      string
	Help      string
	LongHelp  string
	Namespace string
	Disabled  bool
	Labels    []MetricLabel
	TTL       time.Duration
}

// LabelNames returns list of Prometheus labels
func (m *Metric) LabelNames() []string {
	out := make([]string, len(m.Labels))
	for idx, lbl := range m.Labels {
		out[idx] = lbl.Name
	}
	return out
}

var Metrics []Metric
var defaultMetricTTL = 2 * time.Minute

// MetricWithTTL represents the metric with label combinations
// and Last updated time details
type MetricWithTTL struct {
	LastUpdated time.Time
	Labels      prometheus.Labels
}

// ExportedGaugeVec represents each GaugeVec with additional information
type ExportedGaugeVec struct {
	Namespace string
	Name      string
	Help      string
	LongHelp  string
	Labels    []string
	GaugeVec  *prometheus.GaugeVec
	Metrics   map[uint64]MetricWithTTL
	TTL       time.Duration
}

func registerExportedGaugeVec(m Metric, exported *map[string]*ExportedGaugeVec) string {
	gaugeVec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.Namespace,
			Name:      m.Name,
			Help:      m.Help,
		},
		m.LabelNames(),
	)

	// Register the metric with Prometheus
	prometheus.MustRegister(gaugeVec)
	ttl := m.TTL
	if ttl == 0 {
		ttl = defaultMetricTTL
	}

	// Add the metric to the global queue
	Metrics = append(Metrics, m)

	(*exported)[m.Name] = &ExportedGaugeVec{
		Namespace: m.Namespace,
		Name:      m.Name,
		Help:      m.Help,
		LongHelp:  m.LongHelp,
		Labels:    m.LabelNames(),
		GaugeVec:  gaugeVec,
		Metrics:   make(map[uint64]MetricWithTTL),
		TTL:       ttl,
	}
	return m.Name
}

func (gv *ExportedGaugeVec) setMetricLastUpdated(labels prometheus.Labels) {
	if gv.TTL > 0 {
		// Get hash value of Metric labels
		hash := model.LabelsToSignature(labels)
		gv.Metrics[hash] = MetricWithTTL{
			LastUpdated: time.Now(),
			Labels:      labels,
		}
	}
}

// RemoveStaleMetrics removes all the stale metrics which are not
// exported for TTL period.
func (gv *ExportedGaugeVec) RemoveStaleMetrics() {
	if gv.TTL == 0 {
		return
	}

	now := time.Now()
	for _, metric := range gv.Metrics {
		if metric.LastUpdated.Add(gv.TTL).Before(now) {
			gv.GaugeVec.Delete(metric.Labels)
		}
	}
}

// Set updates the Gauge Value and last update time
func (gv *ExportedGaugeVec) Set(labels prometheus.Labels, value float64) {
	gv.GaugeVec.With(labels).Set(value)
	gv.setMetricLastUpdated(labels)
}
