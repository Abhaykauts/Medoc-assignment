package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	baseURL      = "http://localhost:8080"
	doctors      = []string{"doc1", "doc2", "doc3"}
	patientTypes = []string{"EMERGENCY", "PAID_PRIORITY", "FOLLOW_UP", "ONLINE_BOOKING"}
	bookedTokens []string // Store IDs for cancellation test
	mu           sync.Mutex
)

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("🚀 Starting Advanced Multi-Doctor OPD Simulation...")

	// Phase 1: High Concurrency Burst (Morning Rush)
	fmt.Println("\n--- Phase 1: The Morning Rush (Concurrent Bookings) ---")
	runConcurrentRequests(20) // Fire 20 requests at once across 3 docs

	// Phase 2: Slot Saturation & Waitlisting
	fmt.Println("\n--- Phase 2: Stress Testing Waitlist (Filling a Slot) ---")
	fillSlot("doc1", "doc1_slot1") // Ensure this slot is 100% full

	// Phase 3: The Emergency (Preemption)
	fmt.Println("\n--- Phase 3: Emergency Arrives (Testing Preemption) ---")
	// Book Emergency in the FULL slot
	book("doc1", "doc1_slot1", "🚨 CRITICAL PATIENT", "EMERGENCY")

	// Phase 4: The Cancellation (Promotion)
	fmt.Println("\n--- Phase 4: Cancellation Event (Testing Promotion) ---")
	if len(bookedTokens) > 0 {
		// Cancel the first booked token we have
		victimID := bookedTokens[0]
		fmt.Printf("❌ Cancelling Token: %s\n", victimID)
		cancel(victimID)
	} else {
		fmt.Println("⚠️ No booked tokens captured to cancel.")
	}

	printSchedule("doc1")
	fmt.Println("\n✅ Simulation Complete. Please verify logs above for BUMPED/WAITING statuses.")
}

func runConcurrentRequests(count int) {
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(id int) {
			defer wg.Done()
			docID := doctors[rand.Intn(len(doctors))]
			slotID := docID + "_slot1" // Improve collision chance by targeting slot1
			pType := patientTypes[rand.Intn(len(patientTypes))]
			name := fmt.Sprintf("Patient-%d", id)

			book(docID, slotID, name, pType)
		}(i)
	}
	wg.Wait()
}

func fillSlot(docID, slotID string) {
	fmt.Printf("Filling %s capacity to force logic...\n", slotID)
	// Capacity is 3. Let's add 4 normal people.
	for i := 0; i < 4; i++ {
		book(docID, slotID, fmt.Sprintf("Filler-%d", i), "ONLINE_BOOKING")
	}
}

func book(docID, slotID, name, pType string) {
	reqBody, _ := json.Marshal(map[string]string{
		"doctor_id":    docID,
		"slot_id":      slotID,
		"patient_name": name,
		"patient_type": pType,
	})

	resp, err := http.Post(baseURL+"/book", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ Connection Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse status for cleaner logs
	var resMap map[string]interface{}
	json.Unmarshal(body, &resMap)
	status, _ := resMap["status"].(string)
	id, _ := resMap["id"].(string)

	icon := "✅"
	if resp.StatusCode == 409 {
		icon = "❌"
	}
	if status == "WAITING" {
		icon = "⏳"
	}

	fmt.Printf("%s [%s] %s -> %s (Status: %s)\n", icon, docID, pType, name, status)

	if status == "BOOKED" && id != "" {
		mu.Lock()
		bookedTokens = append(bookedTokens, id)
		mu.Unlock()
	}
}

func cancel(tokenID string) {
	reqBody, _ := json.Marshal(map[string]string{
		"token_id": tokenID,
	})

	resp, err := http.Post(baseURL+"/cancel", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ Connection Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ Token Cancelled Successfully! (Check Schedule for Waitlist Promotion)")
	} else {
		fmt.Printf("❌ Cancel Failed: %s\n", resp.Status)
	}
}

func printSchedule(docID string) {
	resp, err := http.Get(fmt.Sprintf("%s/schedule?doctor_id=%s", baseURL, docID))
	if err != nil {
		fmt.Println("Error fetching schedule:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("\n📋 Final Schedule for %s:\n%s\n", docID, string(body))
}
