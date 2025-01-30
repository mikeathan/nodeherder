package automations

import (
	"encoding/json"
	"errors"
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/utils"
	"reflect"
	"sync"
	"time"
)

// brightness + direction_time * value = [expose_name] [+/-] [expose_name] [numeric_operator] [numeric value]
// brightness - direction_time * value

// action set brighness +/- some value = [value_source] [arithmetic operator] [step_value]
// action set brighness  +/- some other numeric combination  eg direction_time * 0.5

const (
	TriggerAction        = "TriggerAction"
	StepAction           = "StepAction"
	PresetRotationAction = "PresetRotationAction"
)

type Step struct {
	Property string `json:"property"`
	Operator string `json:"operator"` // +,-, *,/
	Id       string `json:"id"`
}

type MqttTriggerActionExpose struct {
	Name string `json:"name"`
	Data any    `json:"data,omitempty"`
}

type MqttTriggerAction struct {
	MqttBaseAction
	Exposes []*MqttTriggerActionExpose `json:"exposes"`
	Delay   *utils.TimeInterval        `json:"delay"`
}

func NewTriggerAction() *MqttTriggerAction {
	return &MqttTriggerAction{
		Exposes: make([]*MqttTriggerActionExpose, 0),
		Delay:   utils.IntervalFromMilliseconds(0),
		MqttBaseAction: MqttBaseAction{
			Type: TriggerAction,
		},
	}
}

func (a *MqttTriggerAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {
	bridgeInfo, err := registrar.FindBridgeInfo(a.Id)
	if err != nil {
		return err
	}
	for _, property := range a.Exposes {
		sanitizedData, err := bridgeInfo.SanitiseProperty(property.Name, property.Data)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to sanitize data for action %s: %s", a.Id, err.Error()))
		}
		property.Data = sanitizedData
	}

	a.operation = CreateTriggerOperation(a)

	// common logic
	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}
	// NOTE: if device is renamed we might need to register the automations again
	a.friendlyName = device.FriendlyName
	a.Id = device.Id
	a.Client = client
	a.registrar = registrar
	return nil
}

func (a *MqttTriggerAction) Execute(ctx *DeviceContext) error {
	if a.Delay.Value == 0 {
		return a.executeBase(ctx)
	}

	return a.executeBaseWithDelay(a.Delay, ctx)
}

type MqttStepAction struct {
	MqttBaseAction
	Property string  `json:"property"`
	Steps    []*Step `json:"steps,omitempty"`
	Data     any     `json:"data,omitempty"`
}

func NewStepAction() *MqttStepAction {
	return &MqttStepAction{
		Steps: make([]*Step, 0),
		MqttBaseAction: MqttBaseAction{
			Type: StepAction,
		},
	}
}

func (a *MqttStepAction) Execute(ctx *DeviceContext) error {
	return a.executeBase(ctx)
}

func (a *MqttStepAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}
	expose := device.Exposes[a.Property]
	a.operation = CreateStepOperation(expose, a)

	// common logic
	a.Id = device.Id
	a.friendlyName = device.FriendlyName
	a.Client = client
	a.registrar = registrar

	return nil
}

type MqttPresetCyclingAction struct {
	MqttBaseAction
	Property string   `json:"property"`
	Presets  []string `json:"presets,omitempty"`
}

func NewPresetCyclingAction() *MqttPresetCyclingAction {
	return &MqttPresetCyclingAction{
		Presets: make([]string, 0),
		MqttBaseAction: MqttBaseAction{
			Type: PresetRotationAction,
		},
	}
}

func (a *MqttPresetCyclingAction) Execute(ctx *DeviceContext) error {
	return a.executeBase(ctx)
}

func (a *MqttPresetCyclingAction) Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error {

	device, err := registrar.LookupById(a.Id)
	if err != nil {
		return fmt.Errorf("configure action %s failed: %s ", a.Id, err.Error())
	}

	expose := device.Exposes[a.Property]
	a.operation = CreateRotateOperation(expose)

	// common logic
	a.Id = device.Id
	a.friendlyName = device.FriendlyName
	a.Client = client
	a.registrar = registrar
	return nil
}

type MqttBaseAction struct {
	Id           string                   `json:"id"`
	Type         string                   `json:"type"`
	Client       mqtt.MqttClient          `json:"-"`
	registrar    services.DeviceRegistrar `json:"-"`
	mut          sync.RWMutex             `json:"-"`
	exit         chan bool                `json:"-"`
	isPending    bool                     `json:"-"`
	friendlyName string                   `json:"-"`
	operation    actionOperation          `json:"-"`
}

func (a *MqttBaseAction) emit(payload []byte) {

	msg := fmt.Sprintf("%s/set", a.friendlyName)
	a.Client.Publish(msg, payload)

	utils.LogInfof("Action triggered. Message %s published in %s", string(payload), a.friendlyName)
}

func (b *MqttBaseAction) processAction(ctx *DeviceContext) error {
	payload, err := b.operation.CreatePayload()
	if err != nil {
		return err
	}

	// build payload
	bytes, err := json.Marshal(payload)
	if err != nil {
		utils.LogErrorf("preset cycling action failed %s ", err.Error())
		return err
	}

	// emit message
	b.emit(bytes)

	// on sucess update device context with new values to avoid querying the device again
	// ideally we need to do it if publish has succeeded
	utils.LogInfof("Action triggered. Message %s published in %s", string(bytes), b.friendlyName)
	for key, value := range payload {
		ctx.SetCurrent(key, value)
	}
	return nil
}

func (b *MqttBaseAction) GetID() string {
	return b.Id
}

func (b *MqttBaseAction) GetType() string {
	return b.Type
}

func (a *MqttBaseAction) Stop() {
	if a.isPending {
		// stop it and exit
		a.mut.Lock()
		defer a.mut.Unlock()

		a.exit <- true
		a.isPending = false
	}
}

func (b *MqttBaseAction) executeBaseWithDelay(delay *utils.TimeInterval, ctx *DeviceContext) error {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.isPending {
		return nil
	}

	b.exit = make(chan bool, 1)
	go func() {

		timestamp := time.Now().Add(delay.Duration())
		diff := time.Until(timestamp).Milliseconds()

		duration := time.Duration(diff)
		ticker := *time.NewTicker(duration * time.Millisecond)
		b.isPending = true
		//utils.LogInfof("time constraint started Delay: %d ms", a.Delay)

		defer func() {
			close(b.exit)
			b.isPending = false
		}()

		select {
		case <-ticker.C:

			err := b.processAction(ctx)
			if err != nil {
				utils.LogErrorf("trigger action failed %s", err.Error())
				return
			}
			//utils.LogInfo("timer constraint finished")
			return

		case <-b.exit:

			//utils.LogInfo("timer constraint stopped")
			return
		}
	}()

	return nil
}

func (b *MqttBaseAction) executeBase(ctx *DeviceContext) error {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.isPending {
		return nil
	}

	defer func() {
		b.isPending = false
	}()

	b.isPending = true

	err := b.processAction(ctx)
	if err != nil {
		return err
	}

	return nil
}

type MqttAction interface {
	Execute(tx *DeviceContext) error
	Stop()
	GetID() string
	GetType() string
	Configure(registrar services.DeviceRegistrar, client mqtt.MqttClient) error
}

var typeRegistry = map[string]reflect.Type{
	TriggerAction:        reflect.TypeOf(MqttTriggerAction{}),
	StepAction:           reflect.TypeOf(MqttStepAction{}),
	PresetRotationAction: reflect.TypeOf(MqttPresetCyclingAction{}),
}
