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
- [x] Device disabled feature
- [x] Use device type to identify if diagnostic, feature or expose
- [x] Error reporting
- [x] Metrics repo - keep for X days

### Frontend

- [x] Add app settings in main page
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

```bash
# Test collect endpoint
curl -X POST http://localhost:4100/api/collect \
  -H 'Content-Type: application/json' \
  -d '{"label":"weather node 1","temperature":45.6,"Timestamp":"2023-03-19T19:57:28.961193655Z"}'
```



## 📝 Codebase TODOs

### `./backend/internal/api/types.go`
- [ ] Line 37: implement after fixing the bridge access flatting as now is wrong

### `./backend/internal/automations/device.go`
- [ ] Line 128: can pass the Device event directly

### `./backend/internal/automations/engine.go`
- [ ] Line 16: might need to move it to Models????

### `./backend/internal/automations/scheduler.go`
- [ ] Line 271: for now we assume the automation schedule is in Europe/London timezone

### `./backend/internal/automations/triggers_test.go`
- [ ] Line 630: Empty TODO

### `./backend/internal/controllers/handlers.go`
- [ ] Line 24: maybe do somethng wit the error
- [ ] Line 101: Empty TODO
- [ ] Line 233: if we dont have a request item eg the response came from zigbee2mqtt form their ui

### `./backend/internal/controllers/hub-controller.go`
- [ ] Line 399: see if we can cast p to string and then to bytes
- [ ] Line 450: see if we can cast p to string and then to bytes
- [ ] Line 598: can be refactored to use a factory. for now we will keep it simple
- [ ] Line 635: execute in worker pool
- [ ] Line 647: execute in worker pool

### `./backend/internal/metrics/storage/metrics_test.go`
- [ ] Line 515: Empty TODO

### `./backend/internal/mqtt/mqtt.go`
- [ ] Line 225: remove unused topics if got renamed

### `./backend/internal/services/device_lifetime.go`
- [ ] Line 131: handle this below better
- [ ] Line 162: BUG!
- [ ] Line 251: move it in one place

### `./backend/internal/services/storage_pruning.go`
- [ ] Line 8: Empty TODO

### `./backend/internal/ws/eventhub.go`
- [ ] Line 334: abstract this so we can mock it
- [ ] Line 463: Empty TODO

### `./backend/internal/ws/eventhub_test.go`
- [ ] Line 337: Empty TODO

### `./backend/internal/ws/websocket.go`
- [ ] Line 95: need to segment data if data is too large

### `./backend/mocks/types.go`
- [ ] Line 32: get rid of this. we only used it to have a differnet mocked implementation of Publish

### `./backend/models/devices/device.go`
- [ ] Line 421: make this dynamic

### `./backend/repository/device_file.go`
- [ ] Line 19: use kv database

### `./backend/utils/worker.go`
- [ ] Line 23: maybe do somethng with the error

### `./frontend/src/components/automations/DeviceAutomation.vue`
- [ ] Line 121: ; alert message box to ask user

- [ ] Line 168: find better way to do this we have 2 components that use the same template and toggle from the if isinVieMode

### `./frontend/src/components/automations/Trigger.vue`
- [ ] Line 36: can be refactor to some automation context
- [ ] Line 176: Empty TODO

### `./frontend/src/components/automations/actions/TriggerAction.vue`
- [ ] Line 81: handle more operations when needed
- [ ] Line 99: handle more operations when needed

### `./frontend/src/components/automations/schedule/Scheduler.vue`
- [ ] Line 40: Empty TODO

### `./frontend/src/components/controls/Panel.vue`
- [ ] Line 52: cleanup componentCache ?
- [ ] Line 61: cleanup componentCache ?

### `./frontend/src/components/dashboards/GroupDashboard.vue`
- [ ] Line 90: emit error message
- [ ] Line 98: emit error message
- [ ] Line 118: this is wrong, this should be done in the ws event response in case the request is not successful
- [ ] Line 133: this is wrong, this should be done in the ws event response in case the request is not successful
- [ ] Line 150: this is wrong, this should be done in the ws event response in case the request is not successful

### `./frontend/src/components/device/DeviceExposes.vue`
- [ ] Line 28: do the same we did in Toggle component  so value comes out the correct type eg number

### `./frontend/src/mixins/composables/useAuthentication.ts`
- [ ] Line 18: Empty TODO

### `./frontend/src/modules/formatters/sensor-formatter.ts`
- [ ] Line 205: dont format integer values

### `./frontend/src/transformers/automation/action-transformers.ts`
- [ ] Line 66: add support for icons in operation
