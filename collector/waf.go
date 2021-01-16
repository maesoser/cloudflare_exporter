package collector

func getCloudflareWAFMetrics(startDate string, endDate string, zoneID string, apiEmail string, apiKey string) (respData RespDataStruct, err error) {

	query := `
	{ 
		viewer {
		zones( filter: { zoneTag: $zoneTag } ) {
		  fwEvents: firewallEventsAdaptiveGroups(
			  limit: 5000, 
			  filter: {datetimeMinute_geq: $startDate, datetimeMinute_leq: $endDate}
		  ) {
			count
			dimensions {
			  action
			  clientCountryName
			  clientASNDescription
			  ruleId
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
