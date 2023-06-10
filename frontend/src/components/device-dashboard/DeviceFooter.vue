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
    <span style="margin-left: 0.5rem;margin-right: 3.5rem">
      {{ formatLastSeen(payload) }}
    </span>
    <span :class="`fa fa-fw ${getLinkQualityIcon()}`" style="margin-left: 0.54rem;margin-right: 0.25rem;"></span>
    <span style="margin-right: 0.5rem">
      {{ formatLinkQuality(payload) }}
    </span>
    <span :class="`fa fa-fw ${getBatteryIcon(payload)}`"></span>
  </div>
</template>
