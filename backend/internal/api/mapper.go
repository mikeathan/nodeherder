package api

import "node-herder/models/devices"

 TODO
var SupportedAggregations = map[string][]AggregationType{
	"binary": { AggCount, },
}

func ToDeviceContextResponse(device *devices.Device) *DeviceContextResponse {
	var exposes []ExposeInfo
	for _, exp := range device.Exposes {
		var aggTypes []AggregationType
		for _, agg := range exp.SupportedAggregations {
			aggTypes = append(aggTypes, AggregationType(agg))
		}

		exposes = append(exposes, ExposeInfo{
			Name:         exp.Name,
			Type:         exp.Type,
			Unit:         exp.Unit,
			Values:       exp.Values,
			Aggregations: aggTypes,
		})
	}
}
