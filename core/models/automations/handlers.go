package automations

import (
	"encoding/json"
)

type Handler interface {
	ProcessMessage(payload []byte) error
}

type BridgeHandler struct {
}

func (h *BridgeHandler) ProcessMessage(payload []byte) error {

	//dataMap, err := convertToMap(payload)
	// if err != nil {
	// 	return err
	// }

	// for key, item := range dataMap {

	// 	fmt.Printf("%v =. %v \n", key, item)
	// }
	return nil
}

func convertToMap(payload []byte) ([]map[string]interface{}, error) {

	var deviceMap []map[string]interface{}
	err := json.Unmarshal(payload, &deviceMap)
	if err != nil {
		return nil, err
	}
	return deviceMap, nil
}
