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
	tokenMap map[string]*core.Slot    // Active bookings: TokenID -> Slot
	waitlist map[string][]*core.Token // Waitlist: SlotID -> List of tokens
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
		targetSlot.AddToken(newToken)
		e.tokenMap[newToken.ID] = targetSlot
		return newToken, nil
	}

	// 2. Slot Full - Preemption Check
	tokens := targetSlot.Tokens()
	lowestPriority := 1000
	var replaceCandidate *core.Token

	for _, t := range tokens {
		p := t.Type.Priority()
		if p < lowestPriority {
			lowestPriority = p
			replaceCandidate = t
		} else if p == lowestPriority {
			// Tie-breaker: LIFO (Late arrivals get bumped first)
			if t.Timestamp.After(replaceCandidate.Timestamp) {
				replaceCandidate = t
			}
		}
	}

	// Preempt if new token has higher priority
	if pType.Priority() > lowestPriority {
		fmt.Printf("Preempting %s (%s) for %s (%s)\n", replaceCandidate.ID, replaceCandidate.Type, newToken.ID, newToken.Type)

		// Move candidate to waitlist
		replaceCandidate.Status = "WAITING"
		targetSlot.RemoveToken(replaceCandidate.ID)
		delete(e.tokenMap, replaceCandidate.ID)
		e.addToWaitlist(slotID, replaceCandidate)

		// Add new token
		targetSlot.AddToken(newToken)
		e.tokenMap[newToken.ID] = targetSlot
		return newToken, nil
	}

	// 3. Fallback: Add to Waitlist
	fmt.Printf("Slot full. Waitlisting %s (%s)\n", newToken.ID, newToken.Type)
	newToken.Status = "WAITING"
	e.addToWaitlist(slotID, newToken)

	return newToken, nil
}

func (e *InMemoryEngine) CancelToken(tokenID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	slot, ok := e.tokenMap[tokenID]
	if !ok {
		return core.ErrTokenNotFound
	}

	t := slot.RemoveToken(tokenID)
	if t != nil {
		t.Status = "CANCELLED"
		delete(e.tokenMap, tokenID)

		// Fill the gap from waitlist
		e.promoteFromWaitlist(slot)
		return nil
	}

	return core.ErrTokenNotFound
}

// --- Helpers ---

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

	// Find best candidate (Highest Priority, then Earliest Timestamp)
	highestP := -1
	candidateIdx := -1

	for i, t := range list {
		p := t.Type.Priority()
		if p > highestP {
			highestP = p
			candidateIdx = i
		} else if p == highestP {
			// FIFO
			if t.Timestamp.Before(list[candidateIdx].Timestamp) {
				candidateIdx = i
			}
		}
	}

	if candidateIdx != -1 {
		winner := list[candidateIdx]

		// Remove from waitlist (Swap & Pop)
		list[candidateIdx] = list[len(list)-1]
		e.waitlist[slot.ID] = list[:len(list)-1]

		// Promote to Booked
		winner.Status = "BOOKED"
		slot.AddToken(winner)
		e.tokenMap[winner.ID] = slot

		fmt.Printf("Promoted %s (%s) from waitlist\n", winner.ID, winner.Type)
	}
}
