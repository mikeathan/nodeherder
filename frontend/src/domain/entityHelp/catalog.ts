import type { EntityHelpSource } from './types';

/**
 * One catalog explanation.
 *
 * `text` is plain prose (no markup) and `source` points at the official documentation
 * that supports a vendor-specific claim. Generic device-class entries omit the source.
 */
export type EntityHelpEntry = {
  text: string;
  source?: EntityHelpSource;
};

const EVERYTHING_SMART_HOME: EntityHelpSource = {
  label: 'Everything Smart Home documentation',
  url: 'https://docs.everythingsmart.io/s/products/doc/home-assistant-entities-bUzZQxN7md',
};

/**
 * Curated explanations keyed by ESPHome `object_id` (or its normalized display label).
 *
 * The catalog is data, not code: a new device or firmware entity only adds an entry.
 * Wording follows the manufacturer documentation and avoids claims the firmware does
 * not make. Keys must already be normalized — `catalog.test.ts` enforces that.
 */
export const ENTITY_HELP_BY_OBJECT_ID: Readonly<Record<string, EntityHelpEntry>> = {
  apply_update: {
    text: 'Applies the firmware configuration selected on the device, then checks for an update.',
  },
  assume_present_timeout: {
    text: 'How long, in seconds, to hold occupancy after a target is lost without a valid exit through an entry zone. The firmware default is 1800 s (30 minutes).',
    source: EVERYTHING_SMART_HOME,
  },
  entry_exit_enabled: {
    text: 'Enables best-effort entry/exit tracking. When a target disappears near a doorway, the sensor can hold occupancy instead of clearing immediately.',
    source: EVERYTHING_SMART_HOME,
  },
  esp32_led: {
    text: 'The ESP32 onboard status LED. Turn it off to reduce light pollution in a dark room.',
    source: EVERYTHING_SMART_HOME,
  },
  esp_reboot: {
    text: 'Reboots the ESP32. The device reports offline and its controls are unavailable until it reconnects.',
  },
  everything_presence_lite_firmware: {
    text: 'Firmware published by the manufacturer for this device.',
  },
  exit_threshold_pct: {
    text: 'How much movement inside the entry zone is required before a disappearing target counts as a valid exit. Higher values demand more evidence.',
    source: EVERYTHING_SMART_HOME,
  },
  extra_entities_update: {
    text: 'How often optional tracking entities are published. More frequent updates improve responsiveness but increase traffic and can show temporarily out-of-sync values.',
    source: EVERYTHING_SMART_HOME,
  },
  factory_reset_mmwave_sensor: {
    text: 'Restores the mmWave sensor factory settings. Use it only when tracking is misbehaving.',
    source: EVERYTHING_SMART_HOME,
  },
  illuminance_offset: {
    text: 'Calibration offset added to the measured illuminance, in lux. Use it when the reading differs from a known reference.',
    source: EVERYTHING_SMART_HOME,
  },
  mmwave_bluetooth: {
    text: 'Enables Bluetooth on the mmWave sensor. It is normally only needed for firmware updates, so keep it disabled otherwise.',
    source: EVERYTHING_SMART_HOME,
  },
  occupancy_off_delay: {
    text: 'How long occupancy stays on after the last detected motion. The firmware default is 15 s.',
    source: EVERYTHING_SMART_HOME,
  },
  reboot_mmwave_sensor: {
    text: 'Reboots the mmWave sensor without changing its configuration.',
  },
  stale_target_reset: {
    text: 'Clears a target that the sensor keeps tracking after it has left, using the stale target timeout.',
    source: EVERYTHING_SMART_HOME,
  },
  stale_target_reset_timeout: {
    text: 'How long a target must be unreported before a stale-target reset clears it.',
    source: EVERYTHING_SMART_HOME,
  },
  tracking_behaviour: {
    text: 'Selects which target entities the firmware reports. Disabling unused groups reduces traffic.',
    source: EVERYTHING_SMART_HOME,
  },
  update_speed: {
    text: 'How often the mmWave sensor reports. Slower reporting reduces traffic at the cost of responsiveness.',
    source: EVERYTHING_SMART_HOME,
  },
  upside_down_mounting: {
    text: 'Use when the sensor is mounted upside down (USB-C pointing up) so target tracking stays correctly oriented.',
    source: EVERYTHING_SMART_HOME,
  },
};

/**
 * Standards-based explanations keyed by the discovered `device_class`.
 *
 * These apply to any provider's expose, so they add coverage without device-specific
 * entries. Vendor catalogs win because they are checked first.
 */
export const ENTITY_HELP_BY_DEVICE_CLASS: Readonly<Record<string, EntityHelpEntry>> = {
  battery: { text: 'Battery level reported by the device.' },
  connectivity: { text: 'Whether the device currently has a network connection.' },
  current: { text: 'Electrical current measured by the device.' },
  duration: { text: 'A length of time reported by the device.' },
  energy: { text: 'Accumulated energy consumption.' },
  humidity: { text: 'Relative humidity measured by the device.' },
  illuminance: { text: 'Light level measured in lux.' },
  motion: { text: 'Whether motion is currently detected.' },
  occupancy: { text: 'Whether presence is detected in the monitored area.' },
  power: { text: 'Instantaneous power consumption.' },
  restart: { text: 'Reboots the device.' },
  signal_strength: { text: 'Strength of the device wireless signal.' },
  temperature: { text: 'Temperature measured by the device.' },
  update: { text: 'Firmware update state published by the device.' },
  voltage: { text: 'Voltage measured by the device.' },
};
