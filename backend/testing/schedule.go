package utils_test

import (
	"node-herder/internal/automations"
	"time"
)

func CreateTimeSchedulesWithTimeFormat(startAt time.Time, endAt time.Time, timeFormat string) []*automations.TimeSchedule {
	startSchedule := &automations.TimeSchedule{
		StartAt: startAt.Format(timeFormat),
		Type:    automations.EnableScheduleType,
	}

	endSchedule := &automations.TimeSchedule{
		StartAt: endAt.Format(timeFormat),
		Type:    automations.DisableScheduleType,
	}
	return []*automations.TimeSchedule{startSchedule, endSchedule}

}

func CreateTimeSchedule(startAt time.Time) *automations.TimeSchedule {

	return &automations.TimeSchedule{
		StartAt: startAt.Format("15:04:05.000"),
		Type:    automations.EnableScheduleType,
	}
}

func CreateTimeSchedules(startAt time.Time, endAt time.Time) []*automations.TimeSchedule {
	return CreateTimeSchedulesWithTimeFormat(startAt, endAt, "15:04:05.000")
}
