# NodeHerder TODO

## 🚀 High Priority

### Backend

- [ ] Add support for HTTPS and websocket TLS
- [ ] HTTP polling devices support
- [ ] Add more websocket operation responses (success/error)
- [ ] Test automation loading - configureAction for sanitizing numeric type data
- [ ] Backup automations functionality
- [ ] Device lifetime optimization: check if automations or metrics is enabled before sending event
- [ ] MQTT: Exit after timeout if can't connect
- [ ] RemoveDevice Handler - add context request to emit updated deviceList
- [ ] API rate limiter
- [ ] Cache with expiration

### Frontend

- [ ] Add functionality to enable/disable a trigger
- [ ] Add log window in frontend
- [ ] Update frontend store in dashboard groups websocket response for rename/delete (handle unsuccessful requests)
- [ ] Apply same store update pattern to other websocket requests
- [ ] Automation viewer - enable/disable doesn't save update
- [ ] Handle timerange enum colours
- [ ] Create device card view with sensor data and editor/options view

### Features

- [ ] Deploy to Docker

---

## 🐛 Known Bugs

### Non-Bridge Device Registration

When new non-bridge device joins, ID is built on registration which prevents immediate storage in metrics. Need to:

- Store device ID after registration
- Test if device info persists in store after server restart

---

## 💡 Future Improvements

### Backend

- [ ] Add type in device.expose for HTTP data
- [ ] Add refresh functionality to ping MQTT devices on server start (wake devices)
- [ ] Review: Do we need to unsubscribe from removed/renamed topics?
- [ ] Test new logic in RegisterBridge
- [ ] Review: Device config defaults can't override existing device config overrides (decide if needed)
- [ ] Non-bridge devices (HTTP) - needs investigation/testing
- [ ] Check for disabled items in bridge - add if online

### Frontend

- [ ] Add type in device.expose for HTTP data?
- [ ] Tabs - lazy load tab on click (deferred)
- [ ] Toggle for live data (future consideration)
- [ ] Store response in metrics store? (needs architecture review)

### Metrics

- [ ] Consider sampling data for large datasets
- [ ] Index entries using bolt.Bucket.CreateIndex

### Logging

- [ ] Send MQTT message to enable log type from bridge for zigbee2mqtt event logs
- [ ] Add download log file functionality in UI

---

## ✅ Completed

### Backend

- [x] Add build makefile
- [x] Add remove/force remove/block functionality
- [x] Add configure exposes device functionality
- [x] Add Auth0
- [x] Device disabled feature
- [x] Use device type to identify if diagnostic, feature or expose
- [x] Error reporting
- [x] Metrics repo - keep for X days

### Frontend

- [x] Add app settings in main page
- [x] Add navigation for pages - use Vuetify and redesign layout
- [x] Send multiple messages in one MQTT request for same device
- [x] Update icons to match Home Assistant
- [x] Add device list for devices not shown in dashboard
- [x] Add device groups to be shown in dashboard
- [x] Create defaults for device settings to avoid repetition
- [x] Manage dialogs via event messages
- [x] Add expose selection dialog with multiple selection
- [x] Remove non-measurement exposes from metrics
- [x] Automation schedule - disable/enable button according to scheduler for Manual Trigger
- [x] Device settings component
- [x] Test metrics graph with mocked data in test server
- [x] Metrics results include from/to property for date range
- [x] Device config overrides - fix save on first create

### Features

- [x] Add groups (living room with grouped devices)

---

## 🧪 Test Commands

````bash
```bash
# Test collect endpoint
curl -X POST http://192.168.50.69:4100/api/collect \
  -H 'Content-Type: application/json' \
  -d '{"label":"weather node 1","temperature":45.6,"Timestamp":"2023-03-19T19:57:28.961193655Z"}'
````

```

```
