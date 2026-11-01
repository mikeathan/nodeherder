<script setup lang="ts">
  /* Static contextual help for each spatial tool; labels live in the view domain. */
  import { SPATIAL_CONTEXT_HELP_LABELS, SpatialHelpTopics, type SpatialHelpTopic } from '@/domain/spatial/view';
  import SpatialHelpIllustration from './SpatialHelpIllustration.vue';

  defineProps<{ topic: SpatialHelpTopic }>();
</script>

<template>
  <details class="nh-context-help">
    <summary>{{ SPATIAL_CONTEXT_HELP_LABELS[topic] }}</summary>
    <SpatialHelpIllustration :topic="topic" />
    <div v-if="topic === SpatialHelpTopics.Zones">
      <h3>Creating your first zone</h3>
      <ol>
        <li>Configure target mapping and check a target moves correctly on the Live dashboard.</li>
        <li>
          Choose <strong>Zone editor</strong>, then <strong>Add detection zone</strong>. Select Rectangle for a simple
          area, or Polygon for an irregular shape.
        </li>
        <li>
          Name the zone, then drag it on the canvas. Use the position/dimension fields for precise placement; polygon
          corners have individual X/Y fields. All coordinates are in millimetres from the sensor.
        </li>
        <li>
          Select <strong>Save spatial settings</strong> to store the layout in NodeHerder. These overlays do not
          configure firmware or trigger automations.
        </li>
        <li>
          Walk into the area and compare the geometric preview with the device’s reported occupancy. For a rectangle,
          map its four writable coordinate entities below and select Apply rectangle to device. Use the manufacturer’s
          configurator for polygon firmware zones.
        </li>
      </ol>
      <p>
        Detection zones mark useful areas such as a desk or sofa. Exclusion zones mark false-trigger areas such as
        windows, fans or curtains. Entry zones should span the doorway and extend slightly into the room.
      </p>
      <p>
        The editor offers four detection, two exclusion and two entry slots. Physical support depends on firmware and
        discovered entities. Manufacturer polygon support requires compatible firmware (EPL 1.4.0 or later) and its
        polygon mode switch; drawing an overlay does not enable it.
      </p>
      <p>
        X is left/right (negative/positive); positive Y points away from the sensor. Display orientation rotates the
        view without changing zone coordinates. A mask affects the preview but does not hide raw tracked targets.
      </p>
      <p>
        <a
          href="https://docs.everythingsmart.io/s/products/doc/66bbf380-ac1f-408e-a58b-c878f72783f7"
          target="_blank"
          rel="noopener noreferrer"
          >Manufacturer Zone Editor guide</a
        >
      </p>
    </div>
    <div v-else-if="topic === SpatialHelpTopics.Room">
      <p>
        Start with a measured room outline. Enter 3–16 room-coordinate corners, then add walls, doors and furniture.
        Select an item on the canvas or in the list to edit its position, size and rotation; dragging also moves it.
      </p>
      <p>
        Place the sensor where it is installed and set its facing direction. Targets and zones follow that placement
        together. Room geometry uses millimetres; the sensor remains the origin for zone coordinates.
      </p>
      <p>
        Field of view is an optional illustration that you calibrate from your sensor’s specifications. Set it to zero
        to hide it. Furniture and walls give visual context and do not simulate radio occlusion or alter detection.
      </p>
      <p>
        Use Undo for recent edits. Save spatial settings stores the layout for this device; leaving without saving
        discards drafts. The layout is shared by the dashboard, zone editor and heatmap.
      </p>
      <p>
        <a
          href="https://docs.everythingsmart.io/s/products/doc/room-builder-xrhCh6PsTW"
          target="_blank"
          rel="noopener noreferrer"
          >Manufacturer Room Builder guide</a
        >
      </p>
    </div>
    <div v-else-if="topic === SpatialHelpTopics.Entry">
      <p>
        Everything Presence Lite uses entry zones and recent movement to decide whether a disappearing target probably
        left. If it did not detect a valid exit, firmware can hold occupancy for its configured Assume Present Timeout.
      </p>
      <p>
        This is a best-effort fallback for blind spots or briefly lost stationary targets. It is not a people counter
        and currently exposes no dedicated entry/exit event entities.
      </p>
      <ol>
        <li>Create a simple rectangular Entry overlay spanning the doorway and slightly into the room.</li>
        <li>Configure the corresponding physical entry zone using your device’s supported controls.</li>
        <li>
          Map the reported entities below. Enable Entry / Exit and tune timeout and threshold using their discovered
          limits.
        </li>
        <li>Check Assumed Present and Remaining while testing movement near the doorway.</li>
      </ol>
      <p>
        A higher exit threshold requires more movement evidence; a longer timeout can keep occupancy on after a missed
        exit. Start with a simple doorway rectangle and check the firmware’s own guidance. Support and limits vary by
        firmware.
      </p>
      <p>
        <a
          href="https://docs.everythingsmart.io/s/products/doc/0383f904-51a7-4f96-8ba5-1b85fbe42f49"
          target="_blank"
          rel="noopener noreferrer"
          >Manufacturer Entry/Exit & Assumed Presence guide</a
        >
      </p>
    </div>
    <div v-else>
      <p>
        Heat shows accumulated target observations in a grid: blue is lower activity, through green and yellow to red
        for higher activity, relative to this view’s busiest cell. Counts are sampled target-seconds, not distinct
        people or occupancy percentages.
      </p>
      <p>
        Collection runs once a second while this Spatial tab is open. Choose the last 5, 15 or 60 minutes; there is no
        historical backfill. Missing/offline reports are not measured zero. Hourly totals use your browser’s local
        timezone.
      </p>
      <p>
        Compare hot spots with seating and desk zones; unexpected heat near windows may suggest reflections. A cold area
        can mean low usage or missing tracking, so verify it by walking there before changing zones.
      </p>
      <p>
        Positions stay in memory in this tab for at most one hour, with at most 3600 frames. Leaving the tab, changing
        mappings/device or clearing discards them. Clear affects only this session; saved room and zones remain.
      </p>
      <p>
        <a
          href="https://docs.everythingsmart.io/s/products/doc/4c2f42ad-2a56-4b1c-afa9-09fb7640c1ab"
          target="_blank"
          rel="noopener noreferrer"
          >Manufacturer Heatmap & Analytics guide</a
        >
      </p>
    </div>
  </details>
</template>

<style scoped>
  .nh-context-help {
    margin-block: 0.75rem;
  }
  summary {
    cursor: pointer;
    padding-block: 0.5rem;
  }
  p,
  li {
    line-height: 1.65;
  }
  ol {
    padding-left: 1.4rem;
  }
</style>
