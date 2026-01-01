package api

import (
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"sort"
	"time"
)

var SupportedAggregations = map[bridge.ExposeDataType][]AggregationType{
	bridge.BinaryDataType:  {AggLast, AggCount},
	bridge.NumericDataType: {AggLast, AggMin, AggMax, AggAvg},
	bridge.EnumDataType:    {AggLast, AggCount},
}

func CreateDeviceContextResponse(devices []*devices.Device) *DeviceContextResponse {
	var deviceContexts []*DeviceContext

	for _, device := range devices {
		ctx := toDeviceContextResponse(device)
		if len(ctx.Exposes) == 0 {
			continue
		}

		deviceContexts = append(deviceContexts, ctx)
	}

	return &DeviceContextResponse{
		Version:     DeviceContextVersion,
		GeneratedAt: time.Now().UTC(),
		Devices:     deviceContexts,
	}
}

func toDeviceContextResponse(device *devices.Device) *DeviceContext {
	var exposes []*ExposeInfo

	for _, exp := range device.Exposes {
		expInfo, ok := toExposeInfo(exp)
		if !ok {
			continue
		}
		exposes = append(exposes, expInfo)
	}

	return &DeviceContext{
		ID:          device.Id,
		Name:        device.FriendlyName,
		Description: device.Description,
		Exposes:     exposes,
	}
}

func toExposeInfo(exp *devices.Entity) (*ExposeInfo, bool) {
	if exp.Category != bridge.MeasurementCategory {
		return nil, false
	}

	aggTypes, ok := SupportedAggregations[exp.Type]
	if !ok {
		return nil, false
	}

	info := &ExposeInfo{
		Name:         exp.Name,
		Type:         exp.Type,
		Unit:         exp.Unit,
		Aggregations: aggTypes,
	}

	if exp.Type == bridge.EnumDataType || exp.Type == bridge.BinaryDataType {
		info.Values = extractStringValues(exp.Values)
	}

	// binary semantics
	if exp.Type == bridge.BinaryDataType {
		if v, ok := exp.Values["on"]; ok {
			info.ValueOn = v
		}
		if v, ok := exp.Values["off"]; ok {
			info.ValueOff = v
		}
		if v, ok := exp.Values["toggle"]; ok {
			info.ValueToggle = v
		}
	}

	return info, true
}

func extractStringValues(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			out = append(out, s)
		} else {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
