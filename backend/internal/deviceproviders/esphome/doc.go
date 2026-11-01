// Package esphome implements the ESPHome native API device provider.
//
// The package owns native protocol mapping, discovery, session lifecycle,
// command encoding and the adapter into the shared device pipeline. It does not
// own confirmed device state, automations or browser contracts; those remain in
// the existing services and controllers. Files are split by responsibility:
//
//	config.go      credential validation
//	constants.go   provider name, entity kinds and bounded limits
//	discovery.go   bounded mDNS candidate discovery
//	entity.go      discovery message -> descriptor translation
//	state.go       state message -> observation translation
//	command.go     requested value -> native command translation
//	session.go     one native connection attempt and its ingress/command workers
//	supervisor.go  reconnect, health and command routing for one device
//	device_sink.go adapter from session callbacks to the shared device pipeline
//	driver.go      construction of provider instances for the manager
package esphome
