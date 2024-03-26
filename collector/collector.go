package collector

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "cloudflare"
)

type metricInfo struct {
	Desc *prometheus.Desc
	Type prometheus.ValueType
}

func addMetric(metrics map[string]metricInfo, submodule string, metricName string, docString string, t prometheus.ValueType, labels []string) {
	key := prometheus.BuildFQName(namespace, submodule, metricName)
	// log.Printf("Registered metric %s with labels %v\n", key, labels)
	metrics[metricName] = metricInfo{
		Desc: prometheus.NewDesc(
			key,
			docString,
			labels,
			nil,
		),
		Type: t,
	}
}

// CloudflareCollector is the structure that stores all the information related to the collector
type CloudflareCollector struct {
	apiKey    string
	apiEmail  string
	dataset   []string
	accountID string
	zoneName  string

	API     *cloudflare.API
	zones   []cloudflare.Zone
	account cloudflare.Account

	startDate string
	endDate   string

	cfMetrics map[string]metricInfo

	mutex sync.Mutex
}

func (collector *CloudflareCollector) updateMetric(metricName string, value float64, labelValues ...string) prometheus.Metric {
	// log.Printf("Processing %s with labels: %v\n", metricName, labelValues)
	metric := collector.cfMetrics[metricName]
	if (metric == metricInfo{}) {
		log.Printf("%s metric it is not defined!\n", metricName)
	}
	return prometheus.MustNewConstMetric(metric.Desc, metric.Type, value, labelValues...)
}

// NewCloudflareCollector returns an initialized Collector.
func New(apiKey, apiMail, AccountID, zoneName, dataset string) *CloudflareCollector {

	c := CloudflareCollector{
		apiKey:    apiKey,
		apiEmail:  apiMail,
		accountID: AccountID,
		zoneName:  zoneName,
		dataset:   strings.Split(dataset, ","),
	}

	c.cfMetrics = make(map[string]metricInfo)

	addMetric(c.cfMetrics, "worker", "cputime", "CPU time consumed by worker", prometheus.GaugeValue, []string{"workerName", "accountName", "percentile"})
	addMetric(c.cfMetrics, "worker", "errors", "Errors trigered by worker", prometheus.GaugeValue, []string{"workerName", "accountName"})
	addMetric(c.cfMetrics, "worker", "requests", "Requests received by worker", prometheus.GaugeValue, []string{"workerName", "accountName"})
	addMetric(c.cfMetrics, "worker", "subrequests", "Subrequests performed by worker", prometheus.GaugeValue, []string{"workerName", "accountName"})

	addMetric(c.cfMetrics, "net", "bits", "Number of bits, labelled per AttackID", prometheus.GaugeValue, []string{"attackID", "accountName", "attackProtocol", "mitigationType", "country", "destinationPort", "attackType"})
	addMetric(c.cfMetrics, "net", "packets", "Number of packets, labelled per AttackID", prometheus.GaugeValue, []string{"attackID", "accountName", "attackProtocol", "mitigationType", "country", "destinationPort", "attackType"})

	addMetric(c.cfMetrics, "waf", "events", "Cloudflare WAF Hits", prometheus.GaugeValue, []string{"as", "country", "action", "ruleID", "zoneName"})

	addMetric(c.cfMetrics, "http", "bytes_by_cache_status", "The total number of processed bytes labelled per cache status", prometheus.GaugeValue, []string{"cacheStatus", "method", "contentType", "country", "zoneName"})
	addMetric(c.cfMetrics, "http", "requests_by_response_code", "The total number of request, labelled per HTTP response codes", prometheus.GaugeValue, []string{"responseCode", "zoneName"})
	addMetric(c.cfMetrics, "http", "requests_by_country", "The total number of request, labeled per Country", prometheus.GaugeValue, []string{"country", "zoneName"})
	addMetric(c.cfMetrics, "http", "bytes_by_country", "The total number of request, labeled per Country", prometheus.GaugeValue, []string{"country", "zoneName"})
	addMetric(c.cfMetrics, "http", "threats_by_country", "The total number of threats, labeled per Country", prometheus.GaugeValue, []string{"country", "zoneName"})
	addMetric(c.cfMetrics, "http", "requests_by_content_type", "The total number of request, labeled per content type", prometheus.GaugeValue, []string{"contentType", "zoneName"})
	addMetric(c.cfMetrics, "http", "bytes_by_content_type", "The total number of bytes, labeled per content type", prometheus.GaugeValue, []string{"contentType", "zoneName"})
	addMetric(c.cfMetrics, "http", "requests_by_ssl_version", "The total number of requests labeled per SSL type", prometheus.GaugeValue, []string{"version", "zoneName"})
	addMetric(c.cfMetrics, "http", "requests_by_http_version", "The total number of requests labeled per HTTP version", prometheus.GaugeValue, []string{"version", "zoneName"})
	addMetric(c.cfMetrics, "http", "total_bytes", "The total number of bytes sent", prometheus.GaugeValue, []string{"zoneName"})
	addMetric(c.cfMetrics, "http", "cached_bytes", "The total number of bytes cached", prometheus.GaugeValue, []string{"zoneName"})
	addMetric(c.cfMetrics, "http", "encrypted_bytes", "The total number of bytes encrypted", prometheus.GaugeValue, []string{"zoneName"})
	addMetric(c.cfMetrics, "http", "total_requests", "The total number of requests served", prometheus.GaugeValue, []string{"zoneName"})
	addMetric(c.cfMetrics, "http", "cached_requests", "The total number of requests cached", prometheus.GaugeValue, []string{"zoneName"})
	addMetric(c.cfMetrics, "http", "encrypted_requests", "The total number of requests encrypted", prometheus.GaugeValue, []string{"zoneName"})

	addMetric(c.cfMetrics, "dns", "total_queries", "DNS query count", prometheus.GaugeValue, []string{"zoneName", "queryName", "queryType", "responseCode", "responseCached", "coloName"})
	addMetric(c.cfMetrics, "dns", "uncached_queries", "DNS uncached query count", prometheus.GaugeValue, []string{"zoneName", "queryName", "queryType", "responseCode", "responseCached", "coloName"})
	addMetric(c.cfMetrics, "dns", "staled_queries", "DNS statled queryy count", prometheus.GaugeValue, []string{"zoneName", "queryName", "queryType", "responseCode", "responseCached", "coloName"})
	addMetric(c.cfMetrics, "dns", "average_response_milliseconds", "DNS average response time", prometheus.GaugeValue, []string{"zoneName", "queryName", "queryType", "responseCode", "responseCached", "coloName"})
	addMetric(c.cfMetrics, "dns", "median_response_milliseconds", "DNS median response time", prometheus.GaugeValue, []string{"zoneName", "queryName", "queryType", "responseCode", "responseCached", "coloName"})
	addMetric(c.cfMetrics, "dns", "90th_response_milliseconds", "DNS 90th percentile response time", prometheus.GaugeValue, []string{"zoneName", "queryName", "queryType", "responseCode", "responseCached", "coloName"})
	addMetric(c.cfMetrics, "dns", "99th__response_milliseconds", "DNS 99th percentile response time", prometheus.GaugeValue, []string{"zoneName", "queryName", "queryType", "responseCode", "responseCached", "coloName"})

	err := c.Validate()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Datasets: %v\n", c.dataset)
	if c.zoneName != "" {
		log.Printf("Zone: %s\n", c.zoneName)
	}

	return &c
}

// Describe describes all the metrics ever exported by the Cloudflare exporter. It
// implements prometheus.Collector.
func (collector *CloudflareCollector) Describe(ch chan<- *prometheus.Desc) {

	for _, m := range collector.cfMetrics {
		ch <- m.Desc
	}

}

// Validate checks the configuration parameters given to the Collector
func (collector *CloudflareCollector) Validate() error {

	if collector.apiKey == "" || collector.apiEmail == "" {
		return errors.New("must provide both api-key and api-email")
	}
	if len(collector.dataset) == 0 {
		collector.dataset = append(collector.dataset, "http")
	}
	if contains(collector.dataset, "net") && collector.accountID == "" {
		return errors.New("you must provide an accountID when exporting network analytics")
	}
	if contains(collector.dataset, "workers") && collector.accountID == "" {
		return errors.New("you must provide an accountID when exporting worker analytics")
	}
	return nil
}

// Collect fetches the stats from Cloudflare zones and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (collector *CloudflareCollector) Collect(ch chan<- prometheus.Metric) {
	collector.mutex.Lock()
	defer collector.mutex.Unlock()

	var err error
	err = collector.login()
	if err != nil {
		log.Println(err)
		return
	}
	if contains(collector.dataset, "net") {
		err = collector.collectNetwork(ch)
		if err != nil {
			log.Println(err)
		}
	}
	if contains(collector.dataset, "http") {
		err = collector.collectHTTP(ch)
		if err != nil {
			log.Println(err)
		}
	}
	if contains(collector.dataset, "waf") {
		err = collector.collectWAF(ch)
		if err != nil {
			log.Println(err)
		}
	}
	if contains(collector.dataset, "workers") {
		err = collector.collectWorkers(ch)
		if err != nil {
			log.Println(err)
		}
	}
	if contains(collector.dataset, "dns") {
		err = collector.collectDNS(ch)
		if err != nil {
			log.Println(err)
		}
	}
}

func (collector *CloudflareCollector) login() error {
	collector.startDate = time.Now().Add(time.Duration(-20) * time.Minute).Format(time.RFC3339)
	collector.endDate = time.Now().Add(time.Duration(-5) * time.Minute).Format(time.RFC3339)

	var err error
	collector.API, err = cloudflare.New(collector.apiKey, collector.apiEmail)
	if err != nil {
		return err
	}
	collector.zones, err = collector.API.ListZones(context.Background())
	if err != nil {
		return err
	}
	if collector.accountID != "" {
		collector.account, _, err = collector.API.Account(context.Background(), collector.accountID)
		if err != nil {
			return err
		}
	}
	return err
}

func (collector *CloudflareCollector) collectDNS(ch chan<- prometheus.Metric) error {
	for _, zone := range collector.zones {
		if zone.Plan.ZonePlanCommon.Name != "Enterprise Website" {
			continue
		}
		if zone.Name != "" && zone.Name != collector.zoneName {
			continue
		}
		log.Printf("Getting DNS metrics for %s from %s to %s \n", zone.Name, collector.startDate, collector.endDate)
		resp, err := getCloudflareDNSMetrics(zone.ID, collector.apiEmail, collector.apiKey, buildDNSQueryOptions(collector.startDate, collector.endDate))
		if err == nil {
			for _, node := range resp.Data {
				ch <- collector.updateMetric("total_queries", node.Metrics[0],
					zone.Name, node.Dimensions[0], node.Dimensions[1], node.Dimensions[2], node.Dimensions[3], node.Dimensions[4])
				ch <- collector.updateMetric("uncached_queries", node.Metrics[1],
					zone.Name, node.Dimensions[0], node.Dimensions[1], node.Dimensions[2], node.Dimensions[3], node.Dimensions[4])
				ch <- collector.updateMetric("staled_queries", node.Metrics[2],
					zone.Name, node.Dimensions[0], node.Dimensions[1], node.Dimensions[2], node.Dimensions[3], node.Dimensions[4])
				ch <- collector.updateMetric("average_response_milliseconds", node.Metrics[3],
					zone.Name, node.Dimensions[0], node.Dimensions[1], node.Dimensions[2], node.Dimensions[3], node.Dimensions[4])
				ch <- collector.updateMetric("median_response_milliseconds", node.Metrics[4],
					zone.Name, node.Dimensions[0], node.Dimensions[1], node.Dimensions[2], node.Dimensions[3], node.Dimensions[4])
				ch <- collector.updateMetric("90th_response_milliseconds", node.Metrics[5],
					zone.Name, node.Dimensions[0], node.Dimensions[1], node.Dimensions[2], node.Dimensions[3], node.Dimensions[4])
				ch <- collector.updateMetric("99th__response_milliseconds", node.Metrics[6],
					zone.Name, node.Dimensions[0], node.Dimensions[1], node.Dimensions[2], node.Dimensions[3], node.Dimensions[4])
			}
		} else {
			log.Println("Fetch failed :", err)
		}
	}
	return nil
}

func (collector *CloudflareCollector) collectHTTP(ch chan<- prometheus.Metric) error {
	for _, zone := range collector.zones {
		if zone.Plan.ZonePlanCommon.Name != "Enterprise Website" {
			continue
		}
		if zone.Name != "" && zone.Name != collector.zoneName {
			continue
		}
		log.Printf("Getting HTTP metrics for %s from %s to %s \n", zone.Name, collector.startDate, collector.endDate)
		resp, err := getCloudflareHTTPMetrics(collector.startDate, collector.endDate, zone.ID, collector.apiEmail, collector.apiKey)
		if err == nil {
			for _, node := range resp.Viewer.Zones[0].Caching {
				ch <- collector.updateMetric("bytes_by_cache_status", float64(node.SumEdgeResponseBytes.EdgeResponseBytes),
					node.Dimensions.CacheStatus, node.Dimensions.HTTPMethod, node.Dimensions.ContentTypeName, node.Dimensions.CountryName, zone.Name)
			}

			RequestsData := resp.Viewer.Zones[0].Requests[0].RequestsData

			ch <- collector.updateMetric("total_bytes", float64(RequestsData.Bytes), zone.Name)
			ch <- collector.updateMetric("cached_bytes", float64(RequestsData.CachedBytes), zone.Name)
			ch <- collector.updateMetric("encrypted_bytes", float64(RequestsData.EncryptedBytes), zone.Name)
			ch <- collector.updateMetric("total_requests", float64(RequestsData.Requests), zone.Name)
			ch <- collector.updateMetric("cached_requests", float64(RequestsData.CachedRequests), zone.Name)
			ch <- collector.updateMetric("encrypted_requests", float64(RequestsData.EncryptedRequests), zone.Name)

			for _, node := range RequestsData.ResponseStatusMap {
				ch <- collector.updateMetric("requests_by_response_code", float64(node.Requests), strconv.Itoa(node.EdgeResponseStatus), zone.Name)
			}

			for _, node := range RequestsData.CountryMap {
				ch <- collector.updateMetric("requests_by_country", float64(node.Requests), node.CountryName, zone.Name)
				ch <- collector.updateMetric("bytes_by_country", float64(node.Bytes), node.CountryName, zone.Name)
				ch <- collector.updateMetric("threats_by_country", float64(node.Threats), node.CountryName, zone.Name)
			}
			for _, node := range RequestsData.ContentTypeMap {
				ch <- collector.updateMetric("requests_by_content_type", float64(node.Requests), node.ContentTypeName, zone.Name)
				ch <- collector.updateMetric("bytes_by_content_type", float64(node.Bytes), node.ContentTypeName, zone.Name)
			}
			for _, node := range RequestsData.ClientSSLMap {
				ch <- collector.updateMetric("requests_by_ssl_version", float64(node.Requests), node.ClientSSLProtocol, zone.Name)
			}
			for _, node := range RequestsData.ClientHTTPVersionMap {
				ch <- collector.updateMetric("requests_by_http_version", float64(node.Requests), node.ClientHTTPProtocol, zone.Name)
			}
		} else {
			log.Println("Fetch failed :", err)
		}
	}
	return nil
}

func (collector *CloudflareCollector) collectWAF(ch chan<- prometheus.Metric) error {
	for _, zone := range collector.zones {
		if zone.Plan.ZonePlanCommon.Name != "Enterprise Website" {
			continue
		}
		if zone.Name != "" && zone.Name != collector.zoneName {
			continue
		}
		log.Printf("Getting WAF metrics for %s from %s to %s \n", zone.Name, collector.startDate, collector.endDate)
		resp, err := getCloudflareWAFMetrics(
			collector.startDate, collector.endDate, zone.ID,
			collector.apiEmail,
			collector.apiKey,
		)
		if err == nil {
			for _, node := range resp.Viewer.Zones[0].FwEvents {
				ch <- collector.updateMetric("events", float64(node.Count), node.Dimensions.ASName, node.Dimensions.Country, node.Dimensions.Action, node.Dimensions.RuleID, zone.Name)
			}
		} else {
			log.Println("Fetch failed :", err)
		}
	}
	return nil
}

func (collector *CloudflareCollector) collectWorkers(ch chan<- prometheus.Metric) error {

	log.Printf("Getting Worker metrics for %s from %s to %s \n", collector.accountID, collector.startDate, collector.endDate)
	resp, err := getCloudflareWorkerMetrics(collector.startDate, collector.endDate, collector.accountID, collector.apiEmail, collector.apiKey)
	if err != nil {
		log.Println("Fetch Failed:", err)
		return err
	}
	for _, node := range resp.Viewer.Accounts[0].Workers {
		ch <- collector.updateMetric("cputime", float64(node.Quantiles.CpuTimeP50), node.Info.Name, collector.account.Name, "50")
		ch <- collector.updateMetric("cputime", float64(node.Quantiles.CpuTimeP75), node.Info.Name, collector.account.Name, "75")
		ch <- collector.updateMetric("cputime", float64(node.Quantiles.CpuTimeP99), node.Info.Name, collector.account.Name, "99")
		ch <- collector.updateMetric("cputime", float64(node.Quantiles.CpuTimeP999), node.Info.Name, collector.account.Name, "99.9")
		ch <- collector.updateMetric("errors", float64(node.Sum.Errors), node.Info.Name, collector.account.Name)
		ch <- collector.updateMetric("requests", float64(node.Sum.Requests), node.Info.Name, collector.account.Name)
		ch <- collector.updateMetric("subrequests", float64(node.Sum.SubRequests), node.Info.Name, collector.account.Name)
	}
	return nil
}

func (collector *CloudflareCollector) collectNetwork(ch chan<- prometheus.Metric) error {

	resp, err := getCloudflareNetworkMetrics(collector.startDate, collector.endDate, collector.accountID, collector.apiEmail, collector.apiKey)
	if err == nil {
		for _, node := range resp.Viewer.Accounts[0].NetAttacks {

			ch <- collector.updateMetric("bits",
				float64(node.Sum.Bits),
				node.NetworkDimensions.AttackID,
				collector.account.Name,
				node.NetworkDimensions.AttackProtocol,
				node.NetworkDimensions.AttackMitigationType,
				node.NetworkDimensions.ColoCountry,
				strconv.Itoa(node.NetworkDimensions.DestinationPort),
				node.NetworkDimensions.AttackType,
			)

			ch <- collector.updateMetric("packets",
				float64(node.Sum.Packets),
				node.NetworkDimensions.AttackID,
				collector.account.Name,
				node.NetworkDimensions.AttackProtocol,
				node.NetworkDimensions.AttackMitigationType,
				node.NetworkDimensions.ColoCountry,
				strconv.Itoa(node.NetworkDimensions.DestinationPort),
				node.NetworkDimensions.AttackType,
			)
		}
	} else {
		log.Println("Fetch failed :", err)
	}
	return nil
}

func contains(elements []string, element string) bool {
	for _, e := range elements {
		if element == e {
			return true
		}
	}
	return false
}
