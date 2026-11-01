import type { Device, Expose } from '@/types/device';
import type { SpatialMapping } from '@/types/spatial';
import { exposeDisplayLabel } from '@/domain/exposes';

export type { SpatialMapping } from '@/types/spatial';
export type SpatialTarget = { id: number; x: number; y: number; distance: number };
export type SpatialLabelMismatch = 'axis' | 'target' | 'row' | null;
export type SpatialRole = 'x' | 'y' | 'active';

/** Number of optional target rows shown in the mapping editor. */
export const SPATIAL_TARGET_COUNT = 3;
/** A coordinate older than this is no longer presented as a live position. */
export const SPATIAL_MAX_AGE_MS = 5000;
/** How often freshness is recomputed without a new report. */
export const SPATIAL_FRESHNESS_TICK_MS = 1000;
/** Default graphical range in millimetres. */
export const SPATIAL_DEFAULT_SCALE_MM = 6000;
/** Default view rotation in degrees. */
export const SPATIAL_DEFAULT_ORIENTATION = 0;

/** Selectable graphical range and the label shown for each value. */
export const SPATIAL_SCALE_OPTIONS = [
  { value: 3000, label: '3 m' },
  { value: 6000, label: '6 m' },
  { value: 10000, label: '10 m' },
] as const;

/** Selectable view rotation and the label shown for each value. */
export const SPATIAL_ORIENTATION_OPTIONS = [
  { value: 0, label: 'Forward' },
  { value: 90, label: '90°' },
  { value: 180, label: '180°' },
  { value: 270, label: '270°' },
] as const;

/** Mapping roles in editor display order. */
export const SPATIAL_ROLES: readonly SpatialRole[] = ['x', 'y', 'active'];

/** Length units accepted for spatial coordinates and their millimetre factor. */
export const LENGTH_TO_MM: Readonly<Record<string, number>> = { mm: 1, cm: 10, m: 1000 };

/** Label for one mapping role; "Active" is capitalised, axes stay uppercase. */
export function spatialRoleLabel(role: SpatialRole): string {
  return role === 'active' ? 'Active' : role.toUpperCase();
}

/** Fresh target drafts with every row empty. */
export function createSpatialMappings(): SpatialMapping[] {
  return Array.from({ length: SPATIAL_TARGET_COUNT }, () => ({ x: '', y: '', active: '' }));
}

/** Merges saved rows into a complete editable draft without dropping empty rows. */
export function mergeSpatialMappings(saved: readonly SpatialMapping[] | undefined): SpatialMapping[] {
  return Array.from({ length: SPATIAL_TARGET_COUNT }, (_, index) => ({ x: '', y: '', active: '', ...saved?.[index] }));
}

function numberedLabel(expose: Expose): { target: number; role: SpatialRole } | null {
  if (typeof expose.name !== 'string') return null;
  const match = /^target[\s_-]+(\d{1,4})[\s_-]+(x|y|active)$/i.exec(exposeDisplayLabel(expose).trim());
  return match ? { target: Number(match[1]), role: match[2].toLowerCase() as SpatialRole } : null;
}

/** Prefer matching numbered fields, falling back to unrecognized labels for other devices. */
export function spatialOptions(exposes: Expose[], row: number, role: SpatialRole): Expose[] {
  const matching = exposes.filter((expose) => {
    const label = numberedLabel(expose);
    return label?.target === row && label.role === role;
  });
  return matching.length ? matching : exposes.filter((expose) => !numberedLabel(expose));
}

/** Detects contradictions only in recognizable numbered-target labels; unknown labels remain valid. */
export function spatialLabelMismatch(mapping: SpatialMapping, exposes: Record<string, Expose>, row?: number): SpatialLabelMismatch {
  const targets = new Set<number>();
  for (const [role, id] of Object.entries(mapping)) {
    const expose = exposes[id];
    if (!id || !expose || typeof expose.name !== 'string') continue;
    const label = numberedLabel(expose);
    if (!label) continue;
    if (label.role !== role) return 'axis';
    targets.add(label.target);
  }
  if (targets.size > 1) return 'target';
  if (row !== undefined && targets.size === 1 && !targets.has(row)) return 'row';
  return null;
}

/** Converts a finite coordinate from a declared length unit to millimetres. */
export function coordinateMillimetres(expose: Expose | undefined): number | null {
  if (!expose || typeof expose.data !== 'number' || !Number.isFinite(expose.data)) return null;
  const multiplier = LENGTH_TO_MM[expose.unit?.trim().toLowerCase()];
  return multiplier ? expose.data * multiplier : null;
}

function coordinateIsFresh(id: string, seen: ReadonlyMap<string, number>, now: number): boolean {
  const receivedAt = seen.get(id);
  return receivedAt !== undefined && now >= receivedAt && now - receivedAt <= SPATIAL_MAX_AGE_MS;
}

function mappedTarget(exposes: Record<string, Expose>, mapping: SpatialMapping, index: number, seen: ReadonlyMap<string, number>, now: number): SpatialTarget | null {
  if (spatialLabelMismatch(mapping, exposes, index + 1)) return null;
  if (!mapping.x || !mapping.y || !mapping.active || mapping.x === mapping.y) return null;
  if (!coordinateIsFresh(mapping.x, seen, now) || !coordinateIsFresh(mapping.y, seen, now)) return null;
  if (exposes[mapping.active]?.data !== true) return null;
  const x = coordinateMillimetres(exposes[mapping.x]);
  const y = coordinateMillimetres(exposes[mapping.y]);
  if (x === null || y === null) return null;
  return { id: index + 1, x, y, distance: Math.hypot(x, y) };
}

/** Builds live targets only from explicitly mapped, fresh independent coordinates. */
export function spatialTargets(exposes: Record<string, Expose>, mappings: SpatialMapping[], seen: ReadonlyMap<string, number>, now: number, online: boolean): SpatialTarget[] {
  if (!online) return [];
  const targets: SpatialTarget[] = [];
  for (const [index, mapping] of mappings.entries()) {
    const target = mappedTarget(exposes, mapping, index, seen, now);
    if (target) targets.push(target);
  }
  return targets;
}

/** Why a mapped target row is or is not currently rendered. */
export type SpatialTargetState = 'active' | 'not-active' | 'unknown-active' | 'stale' | 'incomplete' | 'mismatch' | 'offline';
/** One target row's live rendering status. */
export type SpatialTargetStatus = { id: number; state: SpatialTargetState; label: string };

/**
 * Explains each mapping row without changing it. Mapping remains editable while a
 * target is inactive; only live target rendering is suppressed.
 */
export function spatialTargetStatuses(
  exposes: Record<string, Expose>,
  mappings: readonly SpatialMapping[],
  seen: ReadonlyMap<string, number>,
  now: number,
  online: boolean
): SpatialTargetStatus[] {
  return mappings.map((mapping, index) => {
    const id = index + 1;
    if (!mapping.x && !mapping.y && !mapping.active) return { id, state: 'incomplete', label: 'Not mapped' };
    if (!mapping.x || !mapping.y || !mapping.active || mapping.x === mapping.y) {
      return { id, state: 'incomplete', label: 'Incomplete mapping' };
    }
    if (spatialLabelMismatch(mapping, exposes, id)) return { id, state: 'mismatch', label: 'Mapping mismatch' };
    if (!online) return { id, state: 'offline', label: 'Device offline' };
    const active = exposes[mapping.active]?.data;
    if (active !== true) {
      return {
        id,
        state: active === false ? 'not-active' : 'unknown-active',
        label: active === false ? 'Not active' : 'Active state unknown',
      };
    }
    if (!coordinateIsFresh(mapping.x, seen, now) || !coordinateIsFresh(mapping.y, seen, now)) {
      return { id, state: 'stale', label: 'Waiting for fresh X/Y' };
    }
    if (coordinateMillimetres(exposes[mapping.x]) === null || coordinateMillimetres(exposes[mapping.y]) === null) {
      return { id, state: 'stale', label: 'Coordinates unavailable' };
    }
    return { id, state: 'active', label: 'Active' };
  });
}

/** Whether an entity supplies a readable coordinate with a known length unit. */
export function isSpatialCoordinate(entity: Expose): boolean {
  return entity.type === 'numeric' && entity.access_mode !== 'write' && !!LENGTH_TO_MM[(entity.unit || '').trim().toLowerCase()];
}

/** Whether an entity can supply a confirmed active flag. */
export function isSpatialFlag(entity: Expose): boolean {
  return entity.type === 'binary' && entity.access_mode !== 'write';
}

/** Whether a device exposes enough readable entities to offer the spatial view. */
export function supportsSpatialMapping(device: Device | undefined): boolean {
  if (!device?.provider) return false;
  const entities = Object.values(device.exposes ?? {});
  return entities.filter(isSpatialCoordinate).length >= 2 && entities.some(isSpatialFlag);
}

/** Readable coordinate and flag entities for the mapping selectors. */
export function spatialEntities(device: Device): { coordinates: Expose[]; flags: Expose[] } {
  const entities = Object.values(device.exposes ?? {});
  return { coordinates: entities.filter(isSpatialCoordinate), flags: entities.filter(isSpatialFlag) };
}

/** Inventory fields that affect selector identity; live readings are deliberately excluded. */
function spatialInventorySignature(expose: Expose): string {
  return [expose.name, expose.type, expose.unit, expose.access_mode, expose.category, exposeDisplayLabel(expose)].join('\0');
}

/**
 * Reuses the previous selector inventory when only live values changed.
 * Without this, every coordinate report can replace option DOM nodes while the
 * user is choosing, resetting a select back to its previous value.
 */
export function stableSpatialInventory(previous: readonly Expose[] | undefined, next: readonly Expose[]): Expose[] {
  const unchanged =
    previous?.length === next.length &&
    next.every((expose, index) => spatialInventorySignature(expose) === spatialInventorySignature(previous[index]));
  return unchanged ? (previous as Expose[]) : [...next];
}

/** Selector candidates for one target row, grouped by mapping role. */
export type SpatialRowOptions = Record<SpatialRole, Expose[]>;

/** Builds the selector candidates for every editable target row. */
export function spatialRowOptions(mappings: readonly SpatialMapping[], coordinates: Expose[], flags: Expose[]): SpatialRowOptions[] {
  return mappings.map((_, index) => ({
    x: spatialOptions(coordinates, index + 1, 'x'),
    y: spatialOptions(coordinates, index + 1, 'y'),
    active: spatialOptions(flags, index + 1, 'active'),
  }));
}

/** Explains recognized label contradictions; unknown labeling schemes are accepted. */
export function spatialMappingWarning(mapping: SpatialMapping, exposes: Record<string, Expose>, index: number): string {
  const mismatch = spatialLabelMismatch(mapping, exposes, index + 1);
  if (mismatch === 'axis') return 'A selected label does not match its X, Y or Active field.';
  if (mismatch === 'target') return 'X, Y and Active refer to different targets. Choose fields with the same target number.';
  if (mismatch === 'row') return `Choose fields labelled Target ${index + 1} for this row, or use the row matching the selected target.`;
  return '';
}

function mappingError(mapping: SpatialMapping, exposes: Record<string, Expose>, index: number): string {
  if (!mapping.x && !mapping.y && !mapping.active) return '';
  if (!mapping.x || !mapping.y || !mapping.active) return 'choose an X, Y and Active entity, or leave all three blank.';
  if (mapping.x === mapping.y) return 'X and Y must be different entities.';
  const x = exposes[mapping.x], y = exposes[mapping.y], active = exposes[mapping.active];
  if (!x || !y || !active || !isSpatialCoordinate(x) || !isSpatialCoordinate(y) || !isSpatialFlag(active)) {
    return 'a selected entity is no longer available. Choose it again.';
  }
  return spatialMappingWarning(mapping, exposes, index);
}

/** Returns the first actionable mapping error, or an empty string for valid optional targets. */
export function validateSpatialMappings(mappings: SpatialMapping[], exposes: Record<string, Expose>): string {
  for (const [index, mapping] of mappings.entries()) {
    const error = mappingError(mapping, exposes, index);
    if (error) return `Target ${index + 1}: ${error}`;
  }
  return '';
}
