// SPDX-License-Identifier: Apache-2.0

package modules

import "time"

type OnCallUser struct {
	ID          int64
	GitHub      string
	DisplayName string
	Active      bool
	CreatedAt   time.Time
}

type OnCallScheduleRotationPolicy string

const (
	RoundRobinPolicy   OnCallScheduleRotationPolicy = "round-robin"
	SequentialPolicy   OnCallScheduleRotationPolicy = "sequential"
	RandomPolicy       OnCallScheduleRotationPolicy = "random"
)

type OnCallSchedule struct {
	ID                 int64
	Name               string
	Policy             OnCallScheduleRotationPolicy
	Enabled            bool
	CurrentRotationIdx int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type OnCallScheduleUser struct {
	ScheduleID int64
	UserID     int64
	Position   int
}

type OnCallTask struct {
	ID          int64
	ScheduleID  int64
	Repo        string
	IssueNum    int
	Title       string
	Description string
	Status      string
	AssignedTo  int64
	CreatedAt   time.Time
	AckedAt     *time.Time
	CompletedAt *time.Time
}
