package collector

func getCloudflareHTTPMetrics(startDate string, endDate string, zoneID string, apiEmail string, apiKey string) (respData RespDataStruct, err error) {
	query := `
	{
		viewer {
			zones(filter: { zoneTag: $zoneTag }) {
				caching:httpRequestsCacheGroups(
					limit: 10000
					filter: {datetimeMinute_geq: $startDate, datetimeMinute_leq: $endDate}
				) {
					dimensions {
						cacheStatus
						clientCountryName
						clientRequestHTTPMethodName
						edgeResponseContentTypeName
					}
					SumEdgeResponseBytes:sum {	
						edgeResponseBytes	
					}
				}
				requests: httpRequests1mGroups(
					limit: 10000, 
					filter: {datetimeMinute_geq: $startDate, datetimeMinute_leq: $endDate}
				) {
					requestsData:sum {
						bytes
						cachedBytes
						requests
						cachedRequests
						encryptedBytes
						encryptedRequests
						clientSSLMap{
							requests
							clientSSLProtocol
						}
						responseStatusMap{
							edgeResponseStatus
							requests
						}
						clientHTTPVersionMap{
							requests
							clientHTTPProtocol
						}
						contentTypeMap{
							requests
							bytes
							edgeResponseContentTypeName
						}
						countryMap{
							requests
							threats
							clientCountryName
							bytes
						}
					}
				}  
			}
		}
	}
  `
	request := buildGraphQLQuery(query, startDate, endDate, zoneID, "")
	response, err := doGraphQLQuery(request, apiEmail, apiKey)
	return response, err
}
