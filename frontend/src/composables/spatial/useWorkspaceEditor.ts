// Owns selection, undo history and draft mutations for the spatial workspace.
// Persistence is always a separate explicit save from useSpatialView.
import { computed, ref, watch, type Ref } from 'vue';
import type {
  EntryRole,
  RoomObject,
  SpatialPoint,
  SpatialWorkspace,
  SpatialZone,
  FirmwareRole,
  ZoneEntityRole,
  ZoneKind,
  ZoneShape,
} from '@/types/spatial';
import {
  MAX_SPATIAL_OBJECTS,
  MAX_UNDO_SNAPSHOTS,
  STARTER_POLYGON,
  STARTER_ROOM_OBJECT,
  STARTER_ZONE,
  ZONE_LIMITS,
  copyWorkspace,
  defaultZoneName,
  rectanglePoints,
  resizeRectangle,
} from '@/domain/spatial/workspace';
import { createUUID } from '@/utils/unique';

/**
 * Owns selection, draft mutations and a bounded undo stack for the workspace.
 * All writes stay inside the supplied draft; persistence remains an explicit save.
 */
export function useWorkspaceEditor(workspace: Ref<SpatialWorkspace>) {
  const selectedId = ref('');
  let resizeAnchor: SpatialPoint | null = null;
  const undoStack = ref<SpatialWorkspace[]>([]);
  const selectedZone = computed(() => workspace.value.zones.find((zone) => zone.id === selectedId.value));
  const selectedObject = computed(() => workspace.value.objects.find((object) => object.id === selectedId.value));

  watch(workspace, () => {
    resizeAnchor = null;
    if (![...workspace.value.zones, ...workspace.value.objects].some((item) => item.id === selectedId.value)) {
      selectedId.value = '';
    }
    undoStack.value = [];
  });

  /** Captures an undo point before an explicit edit, retaining at most twenty independent snapshots. */
  function checkpoint() {
    undoStack.value = [...undoStack.value.slice(-(MAX_UNDO_SNAPSHOTS - 1)), copyWorkspace(workspace.value)];
  }

  /** Restores the last draft snapshot; persistence requires the normal explicit settings save. */
  function undo() {
    const previous = undoStack.value.pop();
    if (!previous) return;
    // Mutate the current ref value: replacing it would trigger the saved-settings reset watcher.
    Object.assign(workspace.value, previous);
    selectedId.value = '';
  }

  /** Adds an available typed zone slot with starter sensor-relative geometry. */
  function addZone(kind: ZoneKind, shape: ZoneShape = 'rectangle') {
    const count = workspace.value.zones.filter((zone) => zone.kind === kind).length;
    if (count >= ZONE_LIMITS[kind]) return;
    checkpoint();
    const id = createUUID();
    workspace.value.zones.push({
      id,
      name: defaultZoneName(kind, count),
      kind,
      shape,
      points: shape === 'rectangle' ? rectanglePoints(STARTER_ZONE.x, STARTER_ZONE.y, STARTER_ZONE.width, STARTER_ZONE.height) : STARTER_POLYGON.map((point) => ({ ...point })),
    });
    selectedId.value = id;
  }

  /** Adds a bounded visual room item. Millimetres and degrees match persisted room coordinates. */
  function addObject(kind: RoomObject['kind'], name: string = kind) {
    if (workspace.value.objects.length >= MAX_SPATIAL_OBJECTS) return;
    checkpoint();
    const id = createUUID();
    workspace.value.objects.push({
      id,
      name,
      kind,
      x: STARTER_ROOM_OBJECT.x,
      y: STARTER_ROOM_OBJECT.y,
      width: STARTER_ROOM_OBJECT.width,
      height: kind === 'wall' ? STARTER_ROOM_OBJECT.wallHeight : STARTER_ROOM_OBJECT.height,
      rotation: 0,
    });
    selectedId.value = id;
  }

  /** Applies a patch to the selected overlay after capturing one undo point. */
  function updateZone(patch: Partial<SpatialZone>) {
    if (!selectedZone.value) return;
    checkpoint();
    Object.assign(selectedZone.value, patch);
  }

  /** Applies a patch to the selected room item after capturing one undo point. */
  function updateObject(patch: Partial<RoomObject>) {
    if (!selectedObject.value) return;
    checkpoint();
    Object.assign(selectedObject.value, patch);
  }

  /** Updates the sensor transform used to place targets and overlays. */
  function updateSensor(key: keyof SpatialWorkspace['sensor'], value: number) {
    checkpoint();
    workspace.value.sensor[key] = value;
  }

  /** Replaces the sensor pose from a calibration result in one undoable step. */
  function setSensor(patch: Partial<SpatialWorkspace['sensor']>) {
    checkpoint();
    workspace.value.sensor = { ...workspace.value.sensor, ...patch };
  }

  /** Maps one entry/assumed-presence entity to its role. */
  function mapEntry(role: EntryRole, id: string) {
    checkpoint();
    workspace.value.entryEntities[role] = id;
  }

  /** Stores the room outline as editable millimetre points. */
  function setOutline(points: SpatialPoint[]) {
    checkpoint();
    workspace.value.outline = points;
  }

  /** Maps one firmware coordinate role on the selected zone. */
  function mapZone(role: ZoneEntityRole, id: string) {
    if (!selectedZone.value) return;
    updateZone({ entities: { ...selectedZone.value.entities, [role]: id } });
  }

  /** Maps one firmware setting role to an explicitly chosen writable entity. */
  function mapFirmware(role: FirmwareRole, id: string) {
    checkpoint();
    workspace.value.firmwareEntities = { ...(workspace.value.firmwareEntities ?? {}), [role]: id };
  }

  /** Removes the selected overlay only, with immediate undo available until saved/reloaded. */
  function removeSelected() {
    checkpoint();
    workspace.value.zones = workspace.value.zones.filter((zone) => zone.id !== selectedId.value);
    workspace.value.objects = workspace.value.objects.filter((object) => object.id !== selectedId.value);
    selectedId.value = '';
  }

  /** Starts one pointer gesture and captures its undo point and rectangle anchor before any vertex reordering. */
  function beginGesture(index: number | null) {
    checkpoint();
    resizeAnchor = null;
    const zone = selectedZone.value;
    if (zone?.shape === 'rectangle' && index !== null) resizeAnchor = { ...zone.points[(index + 2) % 4] };
  }

  /** Releases the gesture anchor on pointer completion or cancellation. */
  function endGesture() {
    resizeAnchor = null;
  }

  /** Updates a polygon corner or resizes a rectangle against the fixed gesture anchor, in sensor millimetres. */
  function setVertex(index: number, point: SpatialPoint) {
    const zone = selectedZone.value;
    if (!zone) return;
    if (zone.shape === 'polygon') {
      zone.points[index] = { ...point };
      return;
    }
    if (resizeAnchor) zone.points = resizeRectangle(resizeAnchor, point);
  }

  /** Applies a screen-derived millimetre translation to a selected zone or room object. */
  function moveSelected(delta: SpatialPoint) {
    const zone = selectedZone.value;
    const object = selectedObject.value;
    if (zone) zone.points = zone.points.map((point) => ({ x: point.x + delta.x, y: point.y + delta.y }));
    if (object) {
      object.x += delta.x;
      object.y += delta.y;
    }
  }

  return {
    selectedId,
    selectedZone,
    selectedObject,
    undoStack,
    checkpoint,
    undo,
    addZone,
    addObject,
    updateZone,
    updateObject,
    updateSensor,
    setSensor,
    mapEntry,
    mapZone,
    mapFirmware,
    setOutline,
    removeSelected,
    moveSelected,
    beginGesture,
    endGesture,
    setVertex,
  };
}
