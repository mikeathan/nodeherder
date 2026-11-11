package esphome

import "time"

// ProviderName is the persisted provider key and device marker for ESPHome.
const ProviderName = "esphome"

// Native entity kinds. These values are retained alongside a discovered entity
// so a later firmware change cannot silently rebind an identifier to another kind.
const (
	kindSensor       = "sensor"
	kindBinarySensor = "binary_sensor"
	kindTextSensor   = "text_sensor"
	kindText         = "text"
	kindSecretText   = "secret_text"
	kindSwitch       = "switch"
	kindLight        = "light"
	kindNumber       = "number"
	kindSelect       = "select"
	kindButton       = "button"
	kindUnsupported  = "unsupported"
)

// Attribute keys written into the generic entity model. The frontend reads the
// same contract; changing a key requires updating the shared presentation rules.
const (
	attributeLabel              = "label"
	attributeEntityKind         = "entity_kind"
	attributeObjectID           = "object_id"
	attributeDeviceClass        = "device_class"
	attributeControlUnavailable = "control_unavailable"
	attributeMinLength          = "min_length"
	attributeMaxLength          = "max_length"
	attributeMin                = "min"
	attributeMax                = "max"
	attributeStep               = "step"
)

// Wire, storage and resource bounds. Values are deliberately local to this
// integration so a limit change cannot silently affect another provider.
const (
	maxEntities          = 256
	maxEntityNameBytes   = 256
	maxStateTextBytes    = 256
	maxTextValueBytes    = 1024
	maxSelectOptions     = 128
	maxTextPatternBytes  = 256
	commandQueueCapacity = 32
	ingressCapacity      = 512
)

// Session, keepalive and shutdown timing.
const (
	sessionTimeout      = 10 * time.Second
	discoveryTimeout    = 3 * time.Second
	keepaliveInterval   = 20 * time.Second
	closeTimeout        = time.Second
	minReconnectDelay   = time.Second
	maxReconnectDelay   = 60 * time.Second
	stableConnectionAge = time.Minute
)

// Discovery framing bounds. All values are attacker-controlled inputs from the
// local network and must stay bounded before allocation or iteration.
const (
	discoveryPacketBytes   = 9000
	maxDiscoveryRecords    = 256
	maxDiscoveryHosts      = 8
	maxDiscoveryQuestions  = 32
	maxDiscoveryCandidates = 64
)
