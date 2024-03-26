# Cloudflare exporter [![License](https://img.shields.io/github/license/neverlless/cloudflare_exporter)](https://www.gnu.org/licenses/gpl-3.0.html) 

This is a fork of the [Cloudflare Exporter](https://gitlab.com/stephane5/cloudflare_exporter) written by [Stephane](https://gitlab.com/stephane5).

![grafana_image](https://github.com/neverlless/cloudflare_exporter/raw/main/grafana_dashboards/dashboard_image.jpeg)

## Changes from Stephane's version

 - Modified the code structure.
 - Removed [cli](https://github.com/urfave/cli) dependency.
 - Moved the collector to an specific file.
 - You can choose the listening port using `-prom-port`.
 - Now you can add multiple datasets like waf, http, workers, net or dns.
 - If you have several zones but you only want to extract data from one of them you can use `-zone` to specify the **name** of the zone.
 - The metrics are refreshed when the get request is received by prometheus exporter and not a fixed time, following the guidelines by [Prometheus](https://prometheus.io/docs/instrumenting/writing_exporters/#deployment)
 - Modified "http" dataset to return more metrics
 - Added "waf" dataset
 - Added "workers" dataset
 - Added "dns" dataset
 - Added "vdns" dataset

## Supported metrics

- HTTP
   - Bytes (zoneName)
   - Request (zoneName)
   - CachedBytes (zoneName)
   - CachedRequest (zoneName)
   - EncryptedBytes (zoneName)
   - EncryptedRequests (zoneName)

   - Bytes (contentType, zoneName)
   - Request (contentType, zoneName)

   - Bytes (country, zoneName)
   - Request (country, zoneName)
   - Threats (country, zoneName)

   - Bytes (cacheStatus, contentType, method, zoneName)

   - Requests (sslVersion, zoneName)
   - Requests (HTTPVersion, zoneName)
   - Requests (responseCode, zoneName)

- WAF
   - Events (action, asName, country, ruleID, zoneName)

- Workers
   - CPUTime
   - Errors
   - Requests
   - SubRequests

- DNS / DNS Firewall
   - Total Requests
   - Cached Requests
   - Staled Requests
   - Requests time (median,average, 90th percentile, 99th percentile)

- Network
   - Bits (attackID)
   - Packets (attackID)

## Format

Here is a sample of metric you should get once running and fetching from the API

```
cloudflare_total_bytes{zoneName="testdomain.com"} 1.517951e+06
cloudflare_total_requests{zoneName="testdomain.com"} 67
cloudflare_cached_bytes{zoneName="testdomain.com"} 865766
cloudflare_cached_requests{zoneName="testdomain.com"} 15
cloudflare_encrypted_bytes{zoneName="testdomain.com"} 1.502112e+06
cloudflare_encrypted_requests{zoneName="testdomain.com"} 66

cloudflare_requests_per_content_type{contentType="txt",zoneName="testdomain.com"} 3
cloudflare_bytes_per_content_type{contentType="html",zoneName="testdomain.com"} 824591

cloudflare_requests_per_country{country="US",zoneName="testdomain.com"} 33
cloudflare_bytes_per_country{country="DK",zoneName="testdomain.com"} 10514
cloudflare_threats_per_country{country="US",zoneName="testdomain.com"} 31

cloudflare_processed_bytes{cacheStatus="revalidated",contentType="png",method="GET",zoneName="testdomain.com"} 69766

cloudflare_requests_per_http_version{version="TLSv1.3",zoneName="testdomain.com"} 66

cloudflare_requests_per_response_code{responseCode="200",zoneName="testdomain.com"} 9

cloudflare_requests_per_ssl_type{type="HTTP/1.1",zoneName="testdomain.com"} 67

cloudflare_waf_events{action="challenge",as="AS-30083-GO-DADDY-COM-LLC",country="US",ruleID="ip",zoneName="testdomain.com"} 4
```

## Installation

```
go get -u github.com/neverlless/cloudflare_exporter
```

## Usage

```
cloudflare_exporter -h
Usage of ./cloudflare_exporter:
  -account string
    	Account ID to be fetched
  -dataset string
    	The data source you want to export, valid values are: http, net, vdns, dns, workers, waf (default "http,waf")
  -email string
    	The email address associated with your Cloudflare API token and account
  -key string
    	Your Cloudflare API token
  -prom-port string
    	Prometheus Addr (default "0.0.0.0:2112")
  -zone string
    	Zone Name to be fetched
```

You can also use the following env variables instead of cli arguments:
   - `CF_KEY` : Your Cloudflare API token
   - `CF_EMAIL` : The email address associated with your Cloudflare API token and account
   - `CF_ACCOUNT` : Account ID to be fetched
   - `CF_ZONE` : Zone Name to be fetched
   - `CF_DATASET` : The data source you want to export, valid values are: http, net, waf, workers, vnds, dns
   - `CF_PROM_PORT` : Prometheus listening address


Once launched with valid credentials, the binary will spin a webserver on http://localhost:2112/metrics exposing the metrics received from Cloudflare's GraphQL endpoint.

## TODO

- [ ] Add HealthCheck metrics
- [x] Add DNS metrics
- [x] Add DNS Firewall metrics
- [ ] Return old last scrapped metrics if time between scrappings is less than 5 min
- [ ] Refactorize
- [x] Add Grafana Dashboards
- [ ] Add release binaries
