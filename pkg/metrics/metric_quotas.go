package metrics

import (
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	quotaMetricsLabels = []MetricLabel{
		{
			Name: "volume",
			Help: "Name of the volume for which the quota is set",
		},
		{
			Name: "path",
			Help: "Path this under the quota",
		},
	}
	quotasGaugeVers  = make(map[string]*ExportedGaugeVec)
	glusterQuotaUsed = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "quota_used",
		Help:      "Quota used in bytes",
		Labels:    quotaMetricsLabels,
	}, &quotasGaugeVers)
	glusterQuotaAvailable = registerExportedGaugeVec(Metric{
		Namespace: "gluster",
		Name:      "quota_available",
		Help:      "Quota available in bytes",
		Labels:    quotaMetricsLabels,
	}, &quotasGaugeVers)
)

func quotas(gluster glusterutils.GInterface) (err error) {
	qq, err := gluster.Quotas()
	if err != nil {
		log.WithError(err).Debug("[Gluster Quotas] Error:", err)
		return err
	}

	for _, q := range qq {
		labels := prometheus.Labels{
			"volume": q.Volume,
			"path":   q.Path,
		}
		quotasGaugeVers[glusterQuotaUsed].Set(labels, float64(q.Used))
		quotasGaugeVers[glusterQuotaAvailable].Set(labels, float64(q.Available))
	}
	return nil
}

func init() {
	registerMetric("gluster_quotas", quotas)
}
