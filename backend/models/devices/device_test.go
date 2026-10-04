package devices_test

import (
	"encoding/json"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"reflect"
	"sync"
	"testing"
)

func TestEntityDataConcurrentAccess(t *testing.T) {
	data := devices.NewEntityData(float64(0), "brightness")
	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start
			for i := 0; i < 1000; i++ {
				switch worker {
				case 0:
					if i%2 == 0 {
						data.SetValue(float64(i))
					} else {
						data.SetValue("ON")
					}
				case 1:
					wire := "42"
					if i%2 == 0 {
						wire = "true"
					}
					if err := data.UnmarshalJSON([]byte(wire)); err != nil {
						t.Error(err)
					}
				case 2:
					switch data.Value().(type) {
					case float64, string, bool:
					default:
						t.Error("unexpected value type")
					}
					data.ValuesMatch(float64(42))
				case 3:
					encoded, err := json.Marshal(data)
					if err != nil {
						t.Error(err)
						continue
					}
					var value any
					if err := json.Unmarshal(encoded, &value); err != nil {
						t.Errorf("invalid JSON: %s (%v)", encoded, err)
						continue
					}
					switch value := value.(type) {
					case float64:
						if value < 0 || value > 999 {
							t.Errorf("invalid numeric value: %v", value)
						}
					case string:
						if value != "ON" {
							t.Errorf("invalid string: %q", value)
						}
					case bool:
						if !value {
							t.Error("invalid boolean")
						}
					default:
						t.Errorf("unexpected JSON type: %T", value)
					}
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()
}

func TestEntityDataJSONAndComparisonCompatibility(t *testing.T) {
	var zero devices.EntityData
	if zero.Value() != nil {
		t.Fatal("zero value should contain nil")
	}
	for _, tc := range []struct {
		value any
		wire  string
	}{
		{nil, "null"}, {true, "true"}, {"ON", `"ON"`}, {float64(12.5), "12.5"},
		{map[string]any{"x": float64(1)}, `{"x":1}`},
	} {
		zero.SetValue(tc.value)
		encoded, err := json.Marshal(&zero)
		if err != nil || string(encoded) != tc.wire {
			t.Fatalf("JSON = %s (%v), want %s", encoded, err, tc.wire)
		}
		decoded := devices.NewEntityData(nil, "state")
		if err := json.Unmarshal(encoded, decoded); err != nil || !reflect.DeepEqual(decoded.Value(), tc.value) {
			t.Fatalf("round trip = %v (%v), want %v", decoded.Value(), err, tc.value)
		}
	}
	zero.SetValue("retained")
	if err := zero.UnmarshalJSON([]byte("{")); err == nil || zero.Value() != "retained" {
		t.Fatal("invalid JSON changed existing state or succeeded")
	}
	state := devices.NewEntityData("ON", "state")
	if !state.ValuesMatch(true) || state.ValuesMatch(false) {
		t.Fatal("state boolean comparison changed")
	}
	state.SetValue(float64(5))
	if !state.ValuesMatch(float64(5)) {
		t.Fatal("numeric equality changed")
	}
	state.SetValue("ON")
	wire, err := json.Marshal(struct {
		Data *devices.EntityData `json:"data"`
	}{state})
	if err != nil || string(wire) != `{"data":"ON"}` {
		t.Fatalf("nested entity wire shape = %s (%v)", wire, err)
	}
	state.SetValue(func() {})
	if _, err := json.Marshal(state); err == nil {
		t.Fatal("unsupported JSON value unexpectedly succeeded")
	}
	state.SetValue("after encoding error") // encoder errors must not retain the lock
}

type entityReentrantEncoder struct{ data *devices.EntityData }

func (e entityReentrantEncoder) MarshalJSON() ([]byte, error) {
	e.data.SetValue("replacement")
	return []byte(`"encoded"`), nil
}

func TestEntityDataEncoderCanReenter(t *testing.T) {
	data := devices.NewEntityData(nil, "custom")
	data.SetValue(entityReentrantEncoder{data})
	encoded, err := json.Marshal(data)
	if err != nil || string(encoded) != `"encoded"` || data.Value() != "replacement" {
		t.Fatalf("reentrant encoding = %s (%v), value %v", encoded, err, data.Value())
	}
}

func TestEntity_IsEventValue(t *testing.T) {
	tests := []struct {
		name      string
		entity    *devices.Entity
		newValue  any
		wantEvent bool
	}{
		{
			name: "Whitelist property 'action' is always an event",
			entity: &devices.Entity{
				Name: "action",
				Type: bridge.EnumDataType,
			},
			newValue:  "single",
			wantEvent: true,
		},
		{
			name: "Whitelist property 'click' is always an event",
			entity: &devices.Entity{
				Name: "click",
				Type: bridge.EnumDataType,
			},
			newValue:  "double",
			wantEvent: true,
		},
		{
			name: "Schema-defined toggle (string match)",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{
					"toggle": "TOGGLE",
				},
			},
			newValue:  "TOGGLE",
			wantEvent: true,
		},
		{
			name: "Schema-defined toggle (case insensitive match)",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{
					"toggle": "TOGGLE",
				},
			},
			newValue:  "toggle",
			wantEvent: true,
		},
		{
			name: "Schema-defined toggle (type mismatch string vs float64 fallback match)",
			entity: &devices.Entity{
				Name: "brightness",
				Type: bridge.NumericDataType,
				Values: map[string]any{
					"toggle": 1.0,
				},
			},
			newValue:  "1",
			wantEvent: true,
		},
		{
			name: "Legacy toggle command fallback (no schema definition)",
			entity: &devices.Entity{
				Name:   "state",
				Type:   bridge.BinaryDataType,
				Values: map[string]any{},
			},
			newValue:  "TOGGLE",
			wantEvent: true,
		},
		{
			name: "Legacy toggle command fallback case insensitive",
			entity: &devices.Entity{
				Name:   "state",
				Type:   bridge.BinaryDataType,
				Values: map[string]any{},
			},
			newValue:  "toggle",
			wantEvent: true,
		},
		{
			name: "Standard state update (not an event)",
			entity: &devices.Entity{
				Name: "state",
				Type: bridge.BinaryDataType,
				Values: map[string]any{
					"on":  "ON",
					"off": "OFF",
				},
			},
			newValue:  "ON",
			wantEvent: false,
		},
		{
			name: "Standard numeric update (not an event)",
			entity: &devices.Entity{
				Name: "brightness",
				Type: bridge.NumericDataType,
			},
			newValue:  128.0,
			wantEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entity.IsEventValue(tt.newValue); got != tt.wantEvent {
				t.Errorf("Entity.IsEventValue() = %v, want %v", got, tt.wantEvent)
			}
		})
	}
}
