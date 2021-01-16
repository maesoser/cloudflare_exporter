package collector

import (
	"os"
	"testing"
	"time"
)

func TestHTTPMetrics(t *testing.T) {

	var startDate = time.Now().Add(time.Duration(-3) * time.Minute).Format(time.RFC3339)
	var endDate = time.Now().Add(time.Duration(-2) * time.Minute).Format(time.RFC3339)
	_, err := getCloudflareHTTPMetrics(startDate, endDate, "d88b6d7f404e420305cd6c9a73c60576", os.Getenv("APIEMAIL"), os.Getenv("APIKEY"))
	if err != nil {
		t.Errorf("Error: %v", err)
	} else {
		t.Logf("Test succeeded with %v and %v", os.Getenv("apiEmail"), os.Getenv("apiKey"))
	}
}
