package core

import "errors"

var (
	ErrSlotFull          = errors.New("slot is full and no lower priority token found to preempt")
	ErrSlotNotFound      = errors.New("slot not found")
	ErrDoctorNotFound    = errors.New("doctor not found")
	ErrTokenNotFound     = errors.New("token not found")
	ErrInvalidRequest    = errors.New("invalid request")
)

// AllocationEngine defines the core behavior of the OPD system.
type AllocationEngine interface {
	// BookToken attempts to book a token for a patient.
	// Returns the allocated token or an error.
	BookToken(doctorID, slotID string, patientName string, pType PatientType) (*Token, error)

	// CancelToken cancels a specific token.
	// It may trigger reallocation if there's a waiting list (not implemented in this phase yet, but method needed).
	CancelToken(tokenID string) error

	// GetDoctorSchedule returns the doctor with current slots and tokens.
	GetDoctorSchedule(doctorID string) (*Doctor, error)
	
	// AddDoctor adds a doctor to the system (for setup/simulation).
	AddDoctor(doc *Doctor)
}
