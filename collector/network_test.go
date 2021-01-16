package collector

import (
	"log"
	"os"
	"testing"
	"time"
)

func TestNetworkMetrics(t *testing.T) {

	var startDate = time.Now().Add(time.Duration(-3) * time.Minute).Format(time.RFC3339)
	var endDate = time.Now().Add(time.Duration(-2) * time.Minute).Format(time.RFC3339)
	_, err := getCloudflareNetworkMetrics(startDate, endDate, "a63cde259a3885edc49f32101b68379a", os.Getenv("APIEMAIL"), os.Getenv("APIKEY"))
	if err != nil {
		log.Println("Email: ", os.Getenv("APIEMAIL"))
		t.Errorf("Error: %v", err)
	} else {
		t.Logf("Test succeeded with %v and %v", os.Getenv("apiEmail"), os.Getenv("apiKey"))
	}
}
