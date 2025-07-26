TEST:

curl -X POST http://192.168.50.69:4100/collect -H 'Content-Type: application/json' -d '{"label":"weather node 1","temperature":45.6,"Timestamp":"2023-03-19T19:57:28.961193655Z"}'

TODO:

## Deployment

- add build makefile - DONE
- deploy to docker

## Backend

- add type in device.expose for http data
- add refresh functionality to ping mqtt device for when we just started server an we want to awake devices ?
- add remove /force remove/block functionality - DONE
- add configure exposes device functionality - DONE

- add auth0
- add support for https and websocket TLS

- http polling devices support
- add more ws operation responses - eg success or error
- test autiomation loading. configureAction for sanitizing numeric type data
- do we need to unsubsribe from removed/renamed topic ??
- Test new logic in RegisterBridge
- backup automations
- device lifetime optimization : check if automations or metrics is enabled for device before sending event
- device disabled not working !!!!

## Features

- Add schemes - living room with grouped devices

## frontend

- add type in device.expose for http data ?
- add functionality to enable/disable a trigger
- add log window in frontend

BUGS:

- Non bridge new device
  when new nont Bridge device joins
  because it hasnt Id, we build one Id on regisration. So we cant store it in metrics straighr away.
  we would have to set it up afterwards
  Also not sure if server is restarted that we have stored that information eg device id in the store, TO be tested
- Device config overrides on create they dont save the setting first time !!!

- frontend - device settings component - DONE
- frontend - test metrics graph - need mocked data in test node server ! - DONE
- metrics results could have property from/to so we know the range for ui purposes - DONE
- device disabled not working !!!!

-
- Non bridge devices . eg HTTP need more investigation/testing
- error reporting - important - Done
- metrics repo - keep for x days - DONE

frontend - add app settings in main page - DONE
frontend - add navigation for pages - use vuetify and redesign layout - DONE
frontend - Send multiple messages in one mqtt request for same device - DONE

frontend - update icons match homeassistant - DONE
frontend - add device list for devices not shown in dashboad - DONE
frontend - add device groups to be shown in dashboard instead of current dashboard - DONE
frontend/backend - create defauls for some device settings so we dont repeat alot of same info - DONE

TODO:

- frontend -manage the dialogs via event messages - done
- frontend - add expose selection dialog multiple selection - done
- remove non measurement exposes from metrics - done
- automation viewer - enable/disable doesnt save update

- frontend - tabs - load tab on click -(leave for now)
- frontend - handle timerange enum colours
- toggle for live data ? later

once we send the request
store response in metrics store ? needs thinking if we need that

# Logging

send mqqt message to enable log type from bridge to be emmited for zigbee2mqtt event logs
add download file log in UI

Backend TODO

- RemoveDevice Handler add context request so we can emit back the updated deviceList
  api limiter
  cache with expiration
- mqtt: if cant connect after timeout, exit
- use device type to identify if its diagnostic, feature or expose - DONE

METRICS backend TODO

- returns lis of period for ui to choose from - NO
- consider sampling data if too large data set ?
- index entries = bolt.Bucket.CreateIndex

Check for disabled items in bridge - see if we can add them if online
