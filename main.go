package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/neverlless/cloudflare_exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GetEnvStr checks if an environment variable exists and if it does, it returns its value
func GetEnvStr(name, value string) string {
	if os.Getenv(name) != "" {
		return os.Getenv(name)
	}
	return value
}

func main() {
	log.SetPrefix("[cloudflare-exporter] ")
	log.SetFlags(log.Ltime)
	log.SetOutput(os.Stderr)

	APIKey := flag.String("key", GetEnvStr("CF_KEY", ""), "Your Cloudflare API token")
	APIMail := flag.String("email", GetEnvStr("CF_EMAIL", ""), "The email address associated with your Cloudflare API token and account")
	AccountID := flag.String("account", GetEnvStr("CF_ACCOUNT", ""), "Account ID to be fetched")
	zoneName := flag.String("zone", GetEnvStr("CF_ZONE", ""), "Zone Name to be fetched")
	Dataset := flag.String("dataset", GetEnvStr("CF_DATASET", "http,waf"), "The data source you want to export, valid values are: http, network")
	PromListenAddr := flag.String("prom-port", GetEnvStr("CF_PROM_PORT", "0.0.0.0:2112"), "Prometheus Addr")
	flag.Parse()

	CFCollector := collector.New(*APIKey, *APIMail, *AccountID, *zoneName, *Dataset)
	prometheus.MustRegister(CFCollector)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Serving metrics on %s", *PromListenAddr)
	err := http.ListenAndServe(*PromListenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
