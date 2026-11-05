/*
 * Persisted spatial workspace contract. Coordinates are millimetres; sensor
 * rotation and field of view are degrees. These shapes mirror the Go settings
 * model and are validated again at both ingress and persistence boundaries.
 */

/** Supported overlay kinds. */
export const ZONE_KINDS = ['detection', 'exclusion', 'entry'] as const;
/** Supported overlay shapes. */
export const ZONE_SHAPES = ['rectangle', 'polygon'] as const;
/** Writable firmware coordinate roles for a rectangle zone. */
export const ZONE_COORDINATE_ROLES = ['beginX', 'beginY', 'endX', 'endY'] as const;
/** Firmware text entity role for a polygon zone; rectangle coordinates do not apply to it. */
export const ZONE_POLYGON_ROLE = 'polygon';
/** Optional per-zone firmware off-delay entity role. */
export const ZONE_OFF_DELAY_ROLE = 'offDelay';
/** Optional firmware setting roles a workspace may map explicitly. */
export const FIRMWARE_ROLES = [
  'installationAngle',
  'inverseMounting',
  'maxDistance',
  'occupancyOffDelay',
  'polygonMode',
] as const;
/** Supported visual room items. */
export const ROOM_OBJECT_KINDS = ['wall', 'door', 'furniture'] as const;
/** Entry / assumed-presence entity roles. */
export const ENTRY_ROLES = ['enabled', 'timeout', 'threshold', 'assumed', 'remaining'] as const;

/** One optional target mapping: independent X/Y coordinates and a confirmed active flag. */
export type SpatialMapping = { x: string; y: string; active: string };
export type SpatialPoint = { x: number; y: number };
export type ZoneKind = (typeof ZONE_KINDS)[number];
export type ZoneShape = (typeof ZONE_SHAPES)[number];
export type ZoneCoordinateRole = (typeof ZONE_COORDINATE_ROLES)[number];
/** Any firmware entity role a zone can map: four rectangle bounds or one polygon text value. */
export type ZoneEntityRole = ZoneCoordinateRole | typeof ZONE_POLYGON_ROLE | typeof ZONE_OFF_DELAY_ROLE;
/** Firmware settings a workspace can map to writable device entities. */
export type FirmwareRole = (typeof FIRMWARE_ROLES)[number];
export type RoomObjectKind = (typeof ROOM_OBJECT_KINDS)[number];
export type EntryRole = (typeof ENTRY_ROLES)[number];

export type SpatialZone = {
  id: string;
  name: string;
  kind: ZoneKind;
  shape: ZoneShape;
  points: SpatialPoint[];
  entities?: Partial<Record<ZoneEntityRole, string>>;
  occupancyEntity?: string;
};

export type RoomObject = {
  id: string;
  name: string;
  kind: RoomObjectKind;
  x: number;
  y: number;
  width: number;
  height: number;
  rotation: number;
};

export type SpatialWorkspace = {
  version: 1;
  zones: SpatialZone[];
  outline: SpatialPoint[];
  objects: RoomObject[];
  sensor: { x: number; y: number; rotation: number; fov: number };
  entryEntities: Record<EntryRole, string>;
  /** Optional firmware setting mapping; absent for legacy records. */
  firmwareEntities?: Partial<Record<FirmwareRole, string>>;
};

/*
 * Live spatial frame contract (`spatialTargets` WebSocket event). Values are the device's
 * own units and are converted for display; `at` is server time so a client can age a frame
 * without depending on arrival time. Slots are mapping rows, never identities.
 */
export type SpatialFrameAxis = { value: number; unit: string };

export type SpatialFrameTarget = {
  id: number;
  active: boolean;
  x?: SpatialFrameAxis;
  y?: SpatialFrameAxis;
};

export type SpatialFrame = {
  id: string;
  at: string;
  targets: SpatialFrameTarget[];
};
