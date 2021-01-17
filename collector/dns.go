package collector

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type DNSAnalyticsResponse struct {
	Result DNSAnalytics `json:"result"`
}

type DNSAnalytics struct {
	Data []struct {
		Dimensions []string  `json:"dimensions"`
		Metrics    []float64 `json:"metrics"`
	} `json:"data"`
	Totals struct {
		QueryCount         int     `json:"queryCount"`
		ResponseTime90Th   int     `json:"responseTime90th"`
		ResponseTime99Th   int     `json:"responseTime99th"`
		ResponseTimeAvg    float64 `json:"responseTimeAvg"`
		ResponseTimeMedian int     `json:"responseTimeMedian"`
		StaleCount         int     `json:"staleCount"`
		UncachedCount      int     `json:"uncachedCount"`
	} `json:"totals"`
}

func doRequest(url, mail, key string) (respData []byte, err error) {
	client := http.Client{Timeout: time.Second * 5}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-Auth-Key", key)
	request.Header.Set("X-Auth-Email", mail)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return body, nil
}

func buildDNSQueryOptions(startDate, endDate string) string {
	Metrics := []string{"queryCount", "uncachedCount", "staleCount", "responseTimeAvg", "responseTimeMedian", "responseTime90th", "responseTime99th"}
	Dimensions := []string{"queryName", "queryType", "responseCode", "responseCached", "coloName"}
	v := url.Values{}
	v.Set("since", startDate)
	v.Set("until", endDate)
	v.Set("metrics", strings.Join(Metrics, ","))
	v.Set("dimensions", strings.Join(Dimensions, ","))
	return v.Encode()
}

func getCloudflareDNSMetrics(zoneID, mail, key, options string) (respData DNSAnalytics, err error) {
	uri := "https://api.cloudflare.com/client/v4/zones/" + zoneID + "/dns_analytics/report?" + options
	response := DNSAnalyticsResponse{}
	res, err := doRequest(uri, mail, key)
	if err != nil {
		return response.Result, errors.Wrap(err, "error making Request")
	}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return response.Result, errors.Wrap(err, "error making Request")
	}
	return response.Result, nil
}
