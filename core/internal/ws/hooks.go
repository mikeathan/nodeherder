package ws

// type WebSocketHook struct {
// 	wsConn *websocket.Conn
// 	pass func callback instead of using websocket
// }

// func (hook *WebSocketHook) Fire(entry *logrus.Entry) error {
// 	// Convert the log entry to a JSON string
// 	jsonData, err := json.Marshal(entry)
// 	if err != nil {
// 		return err
// 	}

// 	// Send the JSON data as a WebSocket message
// 	err = websocket.WriteJSON(hook.wsConn, jsonData)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
