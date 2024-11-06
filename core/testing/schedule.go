package utils_test

import (
	"node-herder/internal/automations"
	"time"
)

func CreateTimeScheduleWithTimeFormat(startAt time.Time, endAt time.Time, timeFormat string) []*automations.TimeSchedule {
	startSchedule := &automations.TimeSchedule{
		StartAt: startAt.Format(timeFormat),
		Name:    "start job",
		Type:    automations.EnableScheduleType,
	}

	endSchedule := &automations.TimeSchedule{
		StartAt: endAt.Format(timeFormat),
		Name:    "end job",
		Type:    automations.DisableScheduleType,
	}
	return []*automations.TimeSchedule{startSchedule, endSchedule}

}

func CreateTimeSchedule(startAt time.Time, endAt time.Time) []*automations.TimeSchedule {
	return CreateTimeScheduleWithTimeFormat(startAt, endAt, "15:04:05.000")
}
