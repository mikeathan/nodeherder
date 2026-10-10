<script setup lang="ts">
  /* Device summary with live values and controls (Devices screen). */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { RouteName } from '@/types/router';
  import { measurementExposes } from '@/domain/exposes';
  import { isKnownTime } from '@/domain/time';
  import { statusOf } from '@/composables/useHub';
  import { deviceIcon } from '@/components/ui/icons';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import ExposeRow from './ExposeRow.vue';
  import DeviceChips from './DeviceChips.vue';
  import DeviceStatusChip from './DeviceStatusChip.vue';

  const props = defineProps<{ device: Device }>();
  const status = computed(() => statusOf(props.device));
  const rows = computed(() => measurementExposes(props.device).filter((e) => !e.name.startsWith('action')));
</script>

<template>
  <article class="nh-dcard" :class="`is-${status}`" :data-device="device.id">
    <RouterLink :to="{ name: RouteName.DevicePage, params: { id: device.id } }" class="nh-dcard-head">
      <span class="nh-dcard-ic"><UiIcon :path="deviceIcon(device)" /></span>
      <span class="nh-dcard-t">
        <b>{{ device.friendly_name }}</b>
        <small>{{ device.description || device.id }}</small>
      </span>
      <DeviceStatusChip v-if="status !== 'online'" :device="device" />
      <UiIcon name="chevronRight" class="nh-muted" />
    </RouterLink>
    <div v-if="status === 'offline' || status === 'disabled'" class="nh-overlay">
      <UiIcon :name="status === 'offline' ? 'offline' : 'disabled'" />
      <span v-if="status === 'offline'">Offline<template v-if="isKnownTime(device.last_seen)"> · last seen <RelativeTime :value="device.last_seen" /></template></span>
      <span v-else>Disabled in device settings</span>
    </div>
    <div v-else-if="rows.length" class="nh-dcard-rows">
      <ExposeRow v-for="e in rows" :key="e.name" :device="device" :expose="e" />
    </div>
    <p v-else class="nh-dcard-none">No live values. Open the device for details.</p>
    <footer class="nh-dcard-foot">
      <DeviceChips :device="device" />
      <span v-if="isKnownTime(device.last_seen)" class="nh-dcard-seen"><RelativeTime :value="device.last_seen" /></span>
    </footer>
  </article>
</template>

<style scoped>
  .nh-dcard {
    display: flex;
    flex-direction: column;
    background: var(--nh-surface);
    border: var(--nh-border-w) solid transparent;
    border-radius: var(--nh-radius-l);
    box-shadow: var(--nh-shadow-1);
    overflow: hidden;
    min-width: 0;
  }
  [data-mode='dark'] .nh-dcard {
    border-color: var(--nh-border);
  }
  .nh-dcard-head {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 0.8rem var(--nh-space);
    text-decoration: none;
    border-bottom: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-dcard-head:hover {
    background: var(--nh-surface-2);
  }
  .nh-dcard-ic {
    display: grid;
    place-items: center;
    width: 2.3rem;
    height: 2.3rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
    flex: none;
  }
  .nh-dcard-t {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .nh-dcard-t b,
  .nh-dcard-t small {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-dcard-t small {
    color: var(--nh-text-muted);
  }
  .nh-dcard-rows {
    padding: 0.25rem 0;
  }
  .nh-overlay {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.6rem;
    padding: 1.6rem 1rem;
    color: var(--nh-text-muted);
    font-size: 0.875rem;
    text-align: center;
    background: repeating-linear-gradient(-45deg, transparent 0 8px, var(--nh-surface-2) 8px 16px);
  }
  .is-offline .nh-overlay .nh-ic {
    color: var(--nh-danger);
  }
  .nh-dcard-none {
    margin: 0;
    padding: 1rem var(--nh-space);
    color: var(--nh-text-muted);
    font-size: 0.875rem;
  }
  .nh-dcard-foot {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.35rem;
    padding: 0.6rem var(--nh-space);
    border-top: var(--nh-border-w) solid var(--nh-border);
    margin-top: auto;
  }
  .nh-dcard-seen {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--nh-text-muted);
  }
</style>
