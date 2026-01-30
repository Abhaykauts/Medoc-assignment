package algo

import (
	"fmt"
	"medoc-assignment/internal/core"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryEngine implements core.AllocationEngine.
type InMemoryEngine struct {
	doctors  map[string]*core.Doctor
	tokenMap map[string]*core.Slot    // Optimization: TokenID -> Slot (Active bookings only)
	waitlist map[string][]*core.Token // SlotID -> List of waiting tokens
	mu       sync.RWMutex
}

// NewInMemoryEngine creates a new engine instance.
func NewInMemoryEngine() *InMemoryEngine {
	return &InMemoryEngine{
		doctors:  make(map[string]*core.Doctor),
		tokenMap: make(map[string]*core.Slot),
		waitlist: make(map[string][]*core.Token),
	}
}

func (e *InMemoryEngine) AddDoctor(doc *core.Doctor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.doctors[doc.ID] = doc
}

func (e *InMemoryEngine) GetDoctorSchedule(doctorID string) (*core.Doctor, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	doc, ok := e.doctors[doctorID]
	if !ok {
		return nil, core.ErrDoctorNotFound
	}
	return doc, nil
}

func (e *InMemoryEngine) BookToken(doctorID, slotID, patientName string, pType core.PatientType) (*core.Token, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	doc, ok := e.doctors[doctorID]
	if !ok {
		return nil, core.ErrDoctorNotFound
	}

	var targetSlot *core.Slot
	for _, s := range doc.Slots {
		if s.ID == slotID {
			targetSlot = s
			break
		}
	}

	if targetSlot == nil {
		return nil, core.ErrSlotNotFound
	}

	newToken := &core.Token{
		ID:          uuid.New().String(),
		PatientName: patientName,
		Type:        pType,
		Timestamp:   time.Now(),
		SlotID:      slotID,
		DoctorID:    doctorID,
		Status:      "BOOKED",
	}

	// 1. Check Capacity
	if targetSlot.CurrentCount() < targetSlot.Capacity {
		// Space available
		targetSlot.AddToken(newToken)
		e.tokenMap[newToken.ID] = targetSlot
		return newToken, nil
	}

	// 2. Slot Full - Attempt Preemption
	tokens := targetSlot.Tokens()
	lowestPriority := 1000 // Start high
	var replaceCandidate *core.Token

	for _, t := range tokens {
		p := t.Type.Priority()
		if p < lowestPriority {
			lowestPriority = p
			replaceCandidate = t
		} else if p == lowestPriority {
			// Tie-breaking: Bump the one that was created last (LIFO)
			if t.Timestamp.After(replaceCandidate.Timestamp) {
				replaceCandidate = t
			}
		}
	}

	// Check if new token is higher priority
	if pType.Priority() > lowestPriority {
		// PREEMPT: Remove candidate, add new token
		fmt.Printf("Preempting token %s (%s) for new token %s (%s)\n", replaceCandidate.ID, replaceCandidate.Type, newToken.ID, newToken.Type)

		// 1. Update Candidate: Move to Waitlist
		replaceCandidate.Status = "WAITING"
		targetSlot.RemoveToken(replaceCandidate.ID)
		delete(e.tokenMap, replaceCandidate.ID) // Remove from active map

		e.addToWaitlist(slotID, replaceCandidate)

		// 2. Add New Token
		targetSlot.AddToken(newToken)
		e.tokenMap[newToken.ID] = targetSlot

		return newToken, nil
	}

	// 3. Fallback: Add New Token to Waitlist
	fmt.Printf("Slot full. Adding token %s (%s) to waitlist for slot %s\n", newToken.ID, newToken.Type, slotID)
	newToken.Status = "WAITING"
	e.addToWaitlist(slotID, newToken)
	// Note: We do NOT add to tokenMap because currently tokenMap tracks ACTIVE slot tokens.
	// If we want to support cancelling waitlisted tokens, we would need to track them too.
	// For now, adhering to the current pattern.

	return newToken, nil
}

func (e *InMemoryEngine) CancelToken(tokenID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Optimized: Use map lookup
	slot, ok := e.tokenMap[tokenID]
	if !ok {
		return core.ErrTokenNotFound
	}

	// Remove from slot
	t := slot.RemoveToken(tokenID)
	if t != nil {
		t.Status = "CANCELLED"
		delete(e.tokenMap, tokenID)

		// Trigger Reallocation (Fill the gap)
		e.promoteFromWaitlist(slot)
		return nil
	}

	return core.ErrTokenNotFound
}

// Helpers

func (e *InMemoryEngine) addToWaitlist(slotID string, t *core.Token) {
	if _, ok := e.waitlist[slotID]; !ok {
		e.waitlist[slotID] = []*core.Token{}
	}
	e.waitlist[slotID] = append(e.waitlist[slotID], t)
}

func (e *InMemoryEngine) promoteFromWaitlist(slot *core.Slot) {
	list, ok := e.waitlist[slot.ID]
	if !ok || len(list) == 0 {
		return
	}

	// Find highest priority in waitlist
	highestP := -1
	candidateIdx := -1

	for i, t := range list {
		p := t.Type.Priority()
		if p > highestP {
			highestP = p
			candidateIdx = i
		} else if p == highestP {
			// FIFO for waitlist promotion (First come first served among equals)
			// The list is unordered by time, so we should strictly check timestamp if we want strict FIFO
			// But simple index check might suffice if we assumeappend order.
			// Let's explicitly check timestamp for fairness.
			if t.Timestamp.Before(list[candidateIdx].Timestamp) {
				candidateIdx = i
			}
		}
	}

	if candidateIdx != -1 {
		winner := list[candidateIdx]

		// Remove from waitlist
		// Fast remove: swap with last and truncate (order doesn't matter since we scan)
		list[candidateIdx] = list[len(list)-1]
		e.waitlist[slot.ID] = list[:len(list)-1]

		// Add to Slot
		winner.Status = "BOOKED"
		slot.AddToken(winner)
		e.tokenMap[winner.ID] = slot

		fmt.Printf("Promoted token %s (%s) from waitlist to slot %s\n", winner.ID, winner.Type, slot.ID)
	}
}
