package collector

import (
	"context"

	"github.com/machinebox/graphql"
)

type RespDataStruct struct {
	Viewer Viewer `json:"viewer"`
}

type Viewer struct {
	Zones    []Zones   `json:"zones"`
	Accounts []Account `json:"accounts"`
}

type Account struct {
	NetAttacks []AttackHistory `json:"attackHistory"`
	Workers    []Worker        `json:"workers"`
}

type Zones struct {
	Caching  []Caching  `json:"caching"`
	Requests []Requests `json:"requests"`
	FwEvents []FwEvent  `json:"fwEvents"`
}

type Worker struct {
	Info      WorkersInfo      `json:"info"`
	Quantiles WorkersQuantiles `json:"quantiles"`
	Sum       WorkersSum       `json:"sum"`
}

type WorkersInfo struct {
	Name string `json:"scriptName"`
}

type WorkersQuantiles struct {
	CpuTimeP50  float64 `json:"cpuTimeP50"`
	CpuTimeP75  float64 `json:"cpuTimeP75"`
	CpuTimeP99  float64 `json:"cpuTimeP99"`
	CpuTimeP999 float64 `json:"cpuTimeP999"`
}

type WorkersSum struct {
	Errors      int `json:"errors"`
	Requests    int `json:"requests"`
	SubRequests int `json:"subrequests"`
}

type AttackHistory struct {
	NetworkDimensions NetworkDimensions `json:"networkDimensions"`
	Sum               SumAttacks        `json:"sum"`
}

type NetworkDimensions struct {
	AttackID             string `json:"attackId"`
	AttackMitigationType string `json:"attackMitigationType"`
	AttackProtocol       string `json:"attackProtocol"`
	AttackType           string `json:"attackType"`
	ColoCountry          string `json:"coloCountry"`
	DestinationPort      int    `json:"destinationPort"`
}

type SumAttacks struct {
	Bits    int `json:"bits"`
	Packets int `json:"packets"`
}

type Requests struct {
	RequestsData RequestsData `json:"requestsData"`
}

type RequestsData struct {
	Bytes             int `json:"bytes"`
	CachedBytes       int `json:"cachedBytes"`
	EncryptedBytes    int `json:"encryptedBytes"`
	Requests          int `json:"requests"`
	CachedRequests    int `json:"cachedRequests"`
	EncryptedRequests int `json:"encryptedRequests"`

	ResponseStatusMap    []ResponseStatusMap    `json:"responseStatusMap"`
	ClientSSLMap         []ClientSSLMap         `json:"clientSSLMap"`
	ClientHTTPVersionMap []ClientHTTPVersionMap `json:"clientHTTPVersionMap"`
	ContentTypeMap       []ContentTypeMap       `json:"contentTypeMap"`
	CountryMap           []CountryMap           `json:"countryMap"`
}

type Caching struct {
	Dimensions           Dimensions           `json:"dimensions"`
	SumEdgeResponseBytes SumEdgeResponseBytes `json:"sumEdgeResponseBytes"`
}

type Dimensions struct {
	CacheStatus     string `json:"cacheStatus"`
	HTTPMethod      string `json:"clientRequestHTTPMethodName"`
	CountryName     string `json:"clientCountryName"`
	ContentTypeName string `json:"edgeResponseContentTypeName"`
}

type SumEdgeResponseBytes struct {
	EdgeResponseBytes int `json:"edgeResponseBytes"`
}

type ResponseStatusMap struct {
	EdgeResponseStatus int `json:"edgeResponseStatus"`
	Requests           int `json:"requests"`
}

type ClientHTTPVersionMap struct {
	ClientHTTPProtocol string `json:"clientHTTPProtocol"`
	Requests           int    `json:"requests"`
}

type ContentTypeMap struct {
	ContentTypeName string `json:"edgeResponseContentTypeName"`
	Requests        int    `json:"requests"`
	Bytes           int    `json:"bytes"`
}

type ClientSSLMap struct {
	ClientSSLProtocol string `json:"clientSSLProtocol"`
	Requests          int    `json:"requests"`
}

type CountryMap struct {
	CountryName string `json:"clientCountryName"`
	Requests    int    `json:"requests"`
	Bytes       int    `json:"bytes"`
	Threats     int    `json:"threats"`
}

type FwEvent struct {
	Count      int          `json:"count"`
	Dimensions FwDimensions `json:"dimensions"`
}

type FwDimensions struct {
	Action  string `json:"action"`
	ASName  string `json:"clientASNDescription"`
	Country string `json:"clientCountryName"`
	RuleID  string `json:"ruleId"`
}

func buildGraphQLQuery(queryString, startDate, endDate, zoneID, accountID string) *graphql.Request {
	query := graphql.NewRequest(queryString)
	if zoneID != "" {
		query.Var("zoneTag", zoneID)
	}
	if accountID != "" {
		query.Var("accountTag", accountID)
	}
	query.Var("startDate", startDate)
	query.Var("endDate", endDate)
	return query
}

func doGraphQLQuery(query *graphql.Request, apiEmail string, apiKey string) (respData RespDataStruct, err error) {
	client := graphql.NewClient("https://api.cloudflare.com/client/v4/graphql")
	req := query
	req.Header.Set("x-auth-key", apiKey)
	req.Header.Set("x-auth-email", apiEmail)
	ctx := context.Background()
	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, err
	}
	return respData, nil
}
