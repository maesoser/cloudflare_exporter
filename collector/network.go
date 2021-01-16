package collector

func getCloudflareNetworkMetrics(startDate string, endDate string, accountID string, apiEmail string, apiKey string) (respData RespDataStruct, err error) {
	query := `
	{
		networkViewer:viewer {
		  accounts(filter: { accountTag: $accountTag }) {
			attackHistory: ipFlows1mGroups(
			  limit: 10000
			  filter: {datetimeMinute_geq: $startDate, datetimeMinute_leq: $endDate}
			  orderBy: [sum_packets_DESC]
			) {
			  sum {
				bits
				packets
			  }
			  networkDimensions:dimensions {
				attackId
        		coloCountry
        		destinationPort
        		attackType
        		attackMitigationType
        		attackProtocol
			  }
			}
		  }
		}
	  }
	`
	request := buildGraphQLQuery(query, startDate, endDate, "", accountID)
	response, err := doGraphQLQuery(request, apiEmail, apiKey)
	return response, err
}
