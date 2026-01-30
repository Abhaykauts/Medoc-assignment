package core

import "errors"

var (
	ErrSlotFull       = errors.New("slot is full and preemption failed")
	ErrSlotNotFound   = errors.New("slot not found")
	ErrDoctorNotFound = errors.New("doctor not found")
	ErrTokenNotFound  = errors.New("token not found")
	ErrInvalidRequest = errors.New("invalid request")
)

// AllocationEngine defines the core behavior of the allocation system.
type AllocationEngine interface {
	// BookToken reserves a spot for a patient, handling priority logic.
	BookToken(doctorID, slotID string, patientName string, pType PatientType) (*Token, error)

	// CancelToken removes a token and may promote a waiting user.
	CancelToken(tokenID string) error

	// GetDoctorSchedule retrieves the current state of a doctor's schedule.
	GetDoctorSchedule(doctorID string) (*Doctor, error)

	AddDoctor(doc *Doctor)
}
