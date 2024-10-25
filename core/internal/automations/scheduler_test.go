package automations_test

import (
	"fmt"
	"node-herder/internal/automations"
	"testing"
)

func TestAddTimeSchedule(t *testing.T) {

	ts := &automations.TimeSchedule{
		Start: "15:00",
		End:   "16:00",
	}
	s,err := automations.NewScheduler(ts)
	if err != nil {
		t.Errorf("expected %v, got %v", nil, err)
	}
	
	fmt.Println(s)

}
