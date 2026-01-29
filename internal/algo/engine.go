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
	doctors map[string]*core.Doctor
	mu      sync.RWMutex
}

// NewInMemoryEngine creates a new engine instance.
func NewInMemoryEngine() *InMemoryEngine {
	return &InMemoryEngine{
		doctors: make(map[string]*core.Doctor),
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
		return newToken, nil
	}

	// 2. Slot Full - Attempt Preemption
	// Find the token with the lowest priority in the slot.
	// If multiple have the same lowest priority, prefer the one added latest (LIFO for bumping?) or earliest?
	// Usually, we bump the one with lowest priority. If equal, bump the one added latest (LIFO) or keep FIFO?
	// Requirement: "user has higher priority than lowest priority token in slot"

	tokens := targetSlot.Tokens()
	lowestPriority := 1000 // Start high
	var replaceCandidate *core.Token

	for _, t := range tokens {
		p := t.Type.Priority()
		if p < lowestPriority {
			lowestPriority = p
			replaceCandidate = t
		} else if p == lowestPriority {
			// Tie-breaking: Bump the one that was created last? Or first?
			// Let's bump the one created most recently (LIFO) to respect early bookings?
			if t.Timestamp.After(replaceCandidate.Timestamp) {
				replaceCandidate = t
			}
		}
	}

	// Check if new token is higher priority
	if pType.Priority() > lowestPriority {
		// PREEMPT: Remove candidate, add new token
		fmt.Printf("Preempting token %s (%s) for new token %s (%s)\n", replaceCandidate.ID, replaceCandidate.Type, newToken.ID, newToken.Type)

		// Set status of bumped token
		replaceCandidate.Status = "BUMPED"
		// In a real system, we'd add to waitlist here.

		targetSlot.RemoveToken(replaceCandidate.ID)
		targetSlot.AddToken(newToken)

		return newToken, nil
	}

	return nil, core.ErrSlotFull
}

func (e *InMemoryEngine) CancelToken(tokenID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Naive search for token across all doctors/slots
	// Optimized: Maintain a map[tokenID]*Slot

	for _, doc := range e.doctors {
		for _, slot := range doc.Slots {
			if t := slot.RemoveToken(tokenID); t != nil {
				t.Status = "CANCELLED"
				// Here we would check WaitList to fill the gap
				return nil
			}
		}
	}

	return core.ErrTokenNotFound
}
