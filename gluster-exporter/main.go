package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gluster/gluster-prometheus/pkg/conf"
	"github.com/gluster/gluster-prometheus/pkg/doc"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils"
	"github.com/gluster/gluster-prometheus/pkg/glusterutils/glusterconsts"
	"github.com/gluster/gluster-prometheus/pkg/logging"
	"github.com/gluster/gluster-prometheus/pkg/metrics"

	"github.com/Showmax/go-fqdn"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Below variables are set as flags during build time. The current
// values are just placeholders
var (
	exporterVersion         = ""
	defaultGlusterd1Workdir = ""
	defaultGlusterd2Workdir = ""
	defaultConfFile         = ""
)

var (
	showVersion     = flag.Bool("version", false, "Show the version information")
	docgen          = flag.Bool("docgen", false, "Generate exported metrics documentation in Asciidoc format")
	config          = flag.String("config", defaultConfFile, "Config file path")
	defaultInterval = time.Minute
)

func dumpVersionInfo() {
	fmt.Printf("version   : %s\n", exporterVersion)
	fmt.Printf("go version: %s\n", runtime.Version())
	fmt.Printf("go OS/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func getDefaultGlusterdDir(mgmt string) string {
	if mgmt == glusterconsts.MgmtGlusterd2 {
		return defaultGlusterd2Workdir
	}
	return defaultGlusterd1Workdir
}

func main() {
	// Init logger with stderr, will be reinitialized later
	if err := logging.Init("", "-", "info"); err != nil {
		log.Fatal("Init logging failed for stderr")
	}
	flag.Parse()

	if *docgen {
		doc.GenerateMetricsDoc()
		return
	}

	if *showVersion {
		dumpVersionInfo()
		return
	}

	f, err := fqdn.FqdnHostname()
	if err != nil {
		log.WithError(err).Fatal("Failed to guess FQDN")
	}
	metrics.InstanceFQDN = f

	var gluster glusterutils.GInterface
	exporterConf, err := conf.LoadConfig(*config)
	if err != nil {
		log.WithError(err).Fatal("Loading global config failed")
	}

	if strings.ToLower(exporterConf.LogFile) != "stderr" && exporterConf.LogFile != "-" && strings.ToLower(exporterConf.LogFile) != "stdout" {
		// Create Log dir
		err = os.MkdirAll(exporterConf.LogDir, 0750)
		if err != nil {
			log.WithError(err).WithField("logdir", exporterConf.LogDir).
				Fatal("Failed to create log directory")
		}
	}

	if err := logging.Init(exporterConf.LogDir, exporterConf.LogFile, exporterConf.LogLevel); err != nil {
		log.WithError(err).Fatal("Failed to initialize logging")
	}

	// Set the Gluster Configurations used in glusterutils
	if exporterConf.GlusterdWorkdir == "" {
		exporterConf.GlusterdWorkdir =
			getDefaultGlusterdDir(exporterConf.GlusterMgmt)
	}
	// exporter's config will have proper Cluster ID set
	metrics.ClusterID = exporterConf.GlusterClusterID

	gluster = glusterutils.MakeGluster(exporterConf)
	registered := 0
	for _, m := range metrics.GlusterMetrics {
		interval := defaultInterval
		if c, ok := exporterConf.CollectorsConf[m.Name]; ok {
			if c.Disabled {
				continue
			}
			if c.SyncInterval > 0 {
				interval = time.Duration(c.SyncInterval) * time.Second
			}
		}

		go func(m metrics.GlusterMetric, gi glusterutils.GInterface, itvl time.Duration) {
			for {
				if err := m.FN(gi); err != nil {
					log.WithError(err).WithFields(log.Fields{
						"name": m.Name,
					}).Debug("failed to export metric")
				}
				time.Sleep(itvl)
			}
		}(m, gluster, interval)
		registered++
	}

	if registered == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No Metrics registered, Exiting..\n")
		os.Exit(1)
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK\n")) })
	http.Handle(exporterConf.MetricsPath, promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", exporterConf.Port), nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to run exporter\nError: %s", err)
		log.WithError(err).Fatal("Failed to run exporter")
	}
}
