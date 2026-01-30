package core

import (
	"time"
)

// PatientType defines patient priority.
type PatientType string

const (
	Emergency     PatientType = "EMERGENCY"
	PaidPriority  PatientType = "PAID_PRIORITY"
	FollowUp      PatientType = "FOLLOW_UP"
	OnlineBooking PatientType = "ONLINE_BOOKING" // Includes Walk-in
)

// Priority returns the integer priority (higher = more important).
func (pt PatientType) Priority() int {
	switch pt {
	case Emergency:
		return 100
	case PaidPriority:
		return 80
	case FollowUp:
		return 60
	case OnlineBooking:
		return 40
	default:
		return 0
	}
}

// Token represents an allocated or waiting appointment.
type Token struct {
	ID          string      `json:"id"`
	PatientName string      `json:"patient_name"`
	Type        PatientType `json:"type"`
	Timestamp   time.Time   `json:"timestamp"`
	SlotID      string      `json:"slot_id"`
	DoctorID    string      `json:"doctor_id"`
	Status      string      `json:"status"` // BOOKED, CANCELLED, WAITING, BUMPED
}

// Slot represents a discrete time window with fixed capacity.
type Slot struct {
	ID        string    `json:"id"`
	DoctorID  string    `json:"doctor_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Capacity  int       `json:"capacity"`
	tokens    []*Token
}

func NewSlot(id, doctorID string, start, end time.Time, capacity int) *Slot {
	return &Slot{
		ID:        id,
		DoctorID:  doctorID,
		StartTime: start,
		EndTime:   end,
		Capacity:  capacity,
		tokens:    make([]*Token, 0, capacity),
	}
}

// Tokens returns a safe copy of the slot's tokens.
func (s *Slot) Tokens() []*Token {
	result := make([]*Token, len(s.tokens))
	copy(result, s.tokens)
	return result
}

func (s *Slot) AddToken(t *Token) {
	s.tokens = append(s.tokens, t)
}

func (s *Slot) RemoveToken(tokenID string) *Token {
	for i, t := range s.tokens {
		if t.ID == tokenID {
			s.tokens = append(s.tokens[:i], s.tokens[i+1:]...)
			return t
		}
	}
	return nil
}

func (s *Slot) CurrentCount() int {
	return len(s.tokens)
}

// Doctor represents a medical professional and their schedule.
type Doctor struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Slots []*Slot `json:"slots"`
}
