package automations

import "testing"

func TestDeviceInputsIncludesConditionsAndConservativeFallback(t *testing.T) {
	trigger := NewDeviceTrigger("presence")
	trigger.Conditions = []Condition{&ExposeCondition{Name: "illuminance"}, &TimeCondition{}}
	inputs := newTriggerDependencies(TriggerList{trigger})
	for _, property := range []string{"presence", "illuminance"} {
		if !inputs.accepts(map[string]interface{}{property: 1}) {
			t.Fatalf("missing dependency %s", property)
		}
	}
	if inputs.accepts(map[string]interface{}{"linkquality": 134}) {
		t.Fatal("unrelated noise routed")
	}
	trigger.Conditions = append(trigger.Conditions, &BaseCondition{})
	if !newTriggerDependencies(TriggerList{trigger}).accepts(map[string]interface{}{"custom": 1}) {
		t.Fatal("unknown condition must fall back")
	}
	if !newTriggerDependencies(TriggerList{&BaseTrigger{Name: "custom"}}).accepts(map[string]interface{}{"other": 1}) {
		t.Fatal("custom trigger must fall back")
	}
	var unconfigured *triggerDependencies
	if !unconfigured.accepts(nil) {
		t.Fatal("unconfigured instance lost conservative behavior")
	}
}

func BenchmarkDeviceInputsUnrelated(b *testing.B) {
	triggers := make(TriggerList, 100)
	for i := range triggers {
		triggers[i] = NewDeviceTrigger("presence")
	}
	inputs := newTriggerDependencies(triggers)
	noise := map[string]interface{}{"linkquality": 138}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if inputs.accepts(noise) {
			b.Fatal("noise selected")
		}
	}
}
