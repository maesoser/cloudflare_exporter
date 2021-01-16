package collector

func getCloudflareWorkerMetrics(startDate string, endDate string, accountID string, apiEmail string, apiKey string) (respData RespDataStruct, err error) {
	query := `
		{
		Viewer:viewer {
			accounts(filter: {accountTag: $accountTag}) {
			workers:workersInvocationsAdaptive(
				limit: 10000
				filter: {datetimeHour_geq: $startDate, datetimeHour_leq: $endDate }
			) {
				sum {
					subrequests
					requests
					errors
				}
				quantiles {
					cpuTimeP50
					cpuTimeP75
					cpuTimeP99
					cpuTimeP999
				}
				info:dimensions {
					scriptName
				}
			}
			}
		}
		}`

	request := buildGraphQLQuery(query, startDate, endDate, "", accountID)
	response, err := doGraphQLQuery(request, apiEmail, apiKey)
	return response, err
}
