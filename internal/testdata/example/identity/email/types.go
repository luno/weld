package email

import "time"

// TODO(rig): Remove examples.

// Model is a placeholder struct representing the data model
type Model struct {
	ID        int64
	Name      string
	Status    ModelStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ModelStatus int

func (m ModelStatus) ShiftStatus() int {
	return int(m)
}

func (m ModelStatus) ReflexType() int {
	return int(m)
}

const (
	ModelStatusUnknown  ModelStatus = 0
	ModelStatusPending  ModelStatus = 1
	ModelStatusVerified ModelStatus = 2
	modelStatusSentinel ModelStatus = 3 // Should always be last
)

// TODO(rig): Add exported types and enums.
