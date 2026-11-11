import type { EntityHelpSource } from '@/domain/entityHelp';
import type { EntityDefault } from './types';

const EVERYTHING_SMART_HOME: EntityHelpSource = {
  label: 'Everything Smart Home documentation',
  url: 'https://docs.everythingsmart.io/s/products/doc/home-assistant-entities-bUzZQxN7md',
};

const EPL_FIRMWARE: EntityHelpSource = {
  label: 'Everything Presence Lite firmware',
  url: 'https://github.com/EverythingSmartHome/everything-presence-lite/blob/main/common/ld2450-base.yaml',
};

/**
 * Factory defaults keyed by ESPHome `object_id` (or its normalized display label).
 *
 * Every value comes from the firmware YAML (`initial_value`, `initial_option`,
 * `restore_mode`) or the manufacturer documentation. Nothing is inferred from the
 * current state, the minimum or the first option. Keys must already be normalized —
 * `catalog.test.ts` enforces that.
 *
 * Deliberately absent: stateless buttons, unsupported update entities and
 * `mmwave_bluetooth` (toggling it reboots the mmWave sensor).
 */
export const ENTITY_DEFAULTS_BY_OBJECT_ID: Readonly<Record<string, EntityDefault>> = {
  assume_present_timeout: { value: 1800, source: EVERYTHING_SMART_HOME },
  entry_exit_enabled: { value: false, source: EVERYTHING_SMART_HOME },
  esp32_led: { value: false, source: EPL_FIRMWARE },
  exit_threshold_pct: { value: 10, source: EPL_FIRMWARE },
  extra_entities_update: { value: 'Every update', source: EVERYTHING_SMART_HOME },
  illuminance_offset: { value: 0, source: EVERYTHING_SMART_HOME },
  installation_angle: { value: 0, source: EPL_FIRMWARE },
  max_distance: { value: 600, source: EPL_FIRMWARE },
  occupancy_off_delay: { value: 15, source: EVERYTHING_SMART_HOME },
  stale_target_reset: { value: false, source: EVERYTHING_SMART_HOME },
  stale_target_reset_timeout: { value: 30, source: EPL_FIRMWARE },
  tracking_behaviour: { value: 'Above + Speed and Resolution', source: EVERYTHING_SMART_HOME },
  update_speed: { value: 'Normal (0.3s)', source: EVERYTHING_SMART_HOME },
  upside_down_mounting: { value: false, source: EVERYTHING_SMART_HOME },
  zone_1_occupancy_off_delay: { value: 15, source: EPL_FIRMWARE },
  zone_2_occupancy_off_delay: { value: 15, source: EPL_FIRMWARE },
  zone_3_occupancy_off_delay: { value: 15, source: EPL_FIRMWARE },
  zone_4_occupancy_off_delay: { value: 15, source: EPL_FIRMWARE },
};
