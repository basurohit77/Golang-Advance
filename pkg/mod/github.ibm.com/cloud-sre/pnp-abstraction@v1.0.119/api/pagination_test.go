package api

// Moved from PNP Status
import (
	"encoding/json"
	"log"
	"strings"
	"testing"
)

func TestCreatePagination(t *testing.T) {
	log.Print("TestCreatePagination Start")
	inputLimit := 25
	inputOffset := 25
	numberOfRecords := 300
	apiPath := "/path"
	query := "type=incident"
	page := CreatePagination(inputLimit, inputOffset, numberOfRecords, apiPath, query)
	if page.Limit != inputLimit {
		t.Fatal("Limit should be ", inputLimit, ", instead got ", page.Limit)
	}
	log.Print("TestCreatePagination DONE")
}

func TestCreateIBMPagination(t *testing.T) {
	log.Print("TestCreatePagination Start")
	inputLimit := 25
	inputOffset := 25
	numberOfRecords := 300
	apiPath := "/path"
	query := "type=incident"
	page := CreateIBMPagination(inputLimit, inputOffset, numberOfRecords, apiPath, query)
	if page.Count != numberOfRecords {
		t.Fatal("Count should be ", numberOfRecords, ", instead got ", page.Count)
	}
	pageBytes, _ := json.Marshal(page)
	if !strings.Contains(string(pageBytes), "\"total_count\":") {
		t.Fatal("JSON output should contain [\"total_count\":], instead got: ", string(pageBytes))
	}
	log.Print("TestCreateIBMPagination DONE")
}
