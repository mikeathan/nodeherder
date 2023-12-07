package automations_test

import "testing"

func Test1(t *testing.T) {
	var sensorData = buildSensorEvents()
	var automationData = buildBtn1PressReleaseEvents()
	var state []string
	var pos int = 0
	for _, event := range sensorData {

		// need to match in sequence
		// if cache is empty, event need to match with first item of automation

		if event == automationData[pos] {

			if len(automationData) > pos {
				pos++
			}
		}

	}

}
func buildSensorEvents() []string {
	var btn1Events = []string{

		"button_1_press",
		"button_1_press_release"}

	return btn1Events
}
func buildBtn1PressReleaseEvents() []string {
	var btn1Events = []string{

		"button_1_press",
		"button_1_press_release"}

	return btn1Events
}
