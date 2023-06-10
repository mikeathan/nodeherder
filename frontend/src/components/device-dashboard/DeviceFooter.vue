<script setup>
import {
  formatLastSeen,
  getBatteryIcon,
  formatLinkQuality,
  getLinkQualityIcon,
} from "../../modules/sensors/device";

const props = defineProps({
  payload: Object,
});
</script>

<template>
  <div class="card-text">
    <span style="margin-right: 2.5rem">
      {{ formatLastSeen(payload) }}
    </span>
    <span
      :class="`fa fa-fw ${getLinkQualityIcon()}`"
      style="margin-right: 0.25rem !important"
    ></span>
    <span style="margin-right: 0.5rem">
      {{ formatLinkQuality(payload) }}
    </span>
    <span :class="`fa fa-fw ${getBatteryIcon(payload)}`"></span>
  </div>
</template>
