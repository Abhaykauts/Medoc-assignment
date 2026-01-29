package main

import (
	"fmt"
	"log"
	"medoc-assignment/internal/algo"
	"medoc-assignment/internal/api"
	"medoc-assignment/internal/core"
	"net/http"
	"time"
)

func main() {
	// 1. Initialize Engine
	engine := algo.NewInMemoryEngine()

	// 2. Setup Default Doctors (Simulation Data)
	initDoctors(engine)

	// 3. Setup API Handlers
	handler := api.NewHandler(engine)

	http.HandleFunc("/book", handler.BookToken)
	http.HandleFunc("/schedule", handler.GetSchedule)

	port := ":8080"
	fmt.Printf("Starting OPD Token Engine on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func initDoctors(engine core.AllocationEngine) {
	// Create 3 doctors with slots 9-10, 10-11, 11-12
	doctors := []struct {
		ID   string
		Name string
	}{
		{"doc1", "Dr. A (Cardiology)"},
		{"doc2", "Dr. B (Orthopedics)"},
		{"doc3", "Dr. C (General)"},
	}

	now := time.Now()
	// Normalize to today 9:00 AM
	baseTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())

	for _, d := range doctors {
		slots := []*core.Slot{
			core.NewSlot(d.ID+"_slot1", d.ID, baseTime, baseTime.Add(1*time.Hour), 3),                  // 9-10, Cap: 3
			core.NewSlot(d.ID+"_slot2", d.ID, baseTime.Add(1*time.Hour), baseTime.Add(2*time.Hour), 3), // 10-11, Cap: 3
		}

		doc := &core.Doctor{
			ID:    d.ID,
			Name:  d.Name,
			Slots: slots,
		}
		engine.AddDoctor(doc)
	}
}
