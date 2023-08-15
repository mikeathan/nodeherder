package configuration

import (
	"fmt"
	"time"
)

type timerCondition struct {
	Type      string        `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Repeat    bool          `json:"repeat"`
}

type mqttAction struct {
	Friendlyname string `json:"friendlyname"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Value        any    `json:"value"`
}

type trigger struct {
	Type       string            `json:"trigger"`
	Name       string            `json:"name"`
	Conditions []*timerCondition `json:"conditions"`
	Actions    []*mqttAction     `json:"actions"`
}

func newTrigger() *trigger {
	return &trigger{
		Type:       "",
		Name:       "",
		Conditions: []*timerCondition{},
		Actions:    []*mqttAction{},
	}
}

type configuration struct {
	triggers []*trigger
}

func newConfiguration() *configuration {
	return &configuration{
		triggers: []*trigger{},
	}
}

type Factory struct {
}

func Load() *Factory {

	config := createMockConfiguration()
	// TODO: load bridge/devices

	for _, trigger := range config.triggers {
		fmt.Println("loading:", trigger.Name)

	}

	return nil
}

func createMockConfiguration() *configuration {

	trigger := newTrigger()
	trigger.Type = "timer"
	trigger.Name = "Turn on light"

	// new condition
	tc := &timerCondition{}
	tc.Duration = 5 * time.Second
	tc.Repeat = true
	trigger.Conditions = append(trigger.Conditions, tc)

	// new action
	ma := &mqttAction{}
	ma.Friendlyname = "Hive light 1"
	ma.Name = "state"
	ma.Type = "light"
	ma.Value = true

	trigger.Actions = append(trigger.Actions, ma)

	conf := newConfiguration()
	conf.triggers = append(conf.triggers, trigger)
	return conf
}

func Loader() {

}
