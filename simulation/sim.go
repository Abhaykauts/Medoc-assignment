package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// Simulation Request
type BookRequest struct {
	DoctorID    string `json:"doctor_id"`
	SlotID      string `json:"slot_id"`
	PatientName string `json:"patient_name"`
	PatientType string `json:"patient_type"`
}

func main() {
	// Ensure server is running (User must run `go run cmd/service/main.go` separately or we assume it's up)
	// For this script, we'll assume the server is at localhost:8080
	baseURL := "http://localhost:8080"

	fmt.Println("Starting Simulation... (Ensure server is running on :8080)")
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup

	// Sceanrio: 3 Doctors, High Load on Doctor 1
	// Types: EMERGENCY, PAID_PRIORITY, ONLINE_BOOKING

	patientTypes := []string{"ONLINE_BOOKING", "ONLINE_BOOKING", "PAID_PRIORITY", "EMERGENCY"}

	// Simulate 10 concurrent requests to doc1_slot1 (Capacity 3)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			pType := patientTypes[rand.Intn(len(patientTypes))]
			name := fmt.Sprintf("Patient_%d_%s", id, pType)

			req := BookRequest{
				DoctorID:    "doc1",
				SlotID:      "doc1_slot1",
				PatientName: name,
				PatientType: pType,
			}

			statusCode, resp := bookToken(baseURL, req)
			fmt.Printf("[%d] %s -> %d : %s\n", id, pType, statusCode, resp)
		}(i)
	}

	wg.Wait()

	// Fetch Final Schedule
	fmt.Println("\n--- Final Schedule for Doc 1 ---")
	getSchedule(baseURL, "doc1")
}

func bookToken(baseURL string, req BookRequest) (int, string) {
	data, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/book", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Status
}

func getSchedule(baseURL, doctorID string) {
	resp, err := http.Get(baseURL + "/schedule?doctor_id=" + doctorID)
	if err != nil {
		fmt.Println("Error fetching schedule:", err)
		return
	}
	defer resp.Body.Close()

	var doc interface{}
	json.NewDecoder(resp.Body).Decode(&doc)
	formatted, _ := json.MarshalIndent(doc, "", "  ")
	fmt.Println(string(formatted))
}
