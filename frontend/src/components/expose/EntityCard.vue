<script setup lang="ts">
  import { getSensorValue, getSensorIcon, getSensorUnit } from '../../modules/formatters/sensor-formatter';
  import { getEntityIcon } from '../../modules/formatters/entity.formatter';
  import { store } from '../../store/index';
  import { computed, ref } from 'vue';
  import { Device, Expose } from '@/types/device';
  import Icon from '../controls/Icon.vue';
  import { mdiCeilingLightMultiple } from '@mdi/js';

  const props = defineProps({
    id: { type: String, required: true },
    name: { type: String, required: true },
  });

  const expose = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    //if (!device) return null;

    return device.exposes[props.name] as Expose;
  });

  function handleIconClick(): void {
    console.log('handleIconClick');
  }
  const lightOn = ref(true);
</script>

<style scoped>
  .entity-card {
    padding: 1em;
    background: #fff;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
    display: flex;
    flex-direction: column;
    gap: 1em;
  }

  .entity-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
  }

  .entity-labels {
    display: flex;
    flex-direction: column;
    justify-content: center;
  }

  .entity-labels {
    display: flex;
    flex-direction: column;
    justify-content: center;
  }

  .entity-title {
    font-size: 1rem;
    font-weight: 500;
  }

  .entity-value {
    font-size: 0.75rem;
    color: #777;
    margin-top: 0.2em;
  }

  /* .entity-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    cursor: pointer;
    transition: background-color 0.2s;
  } */
</style>
<template>
  <Card>
    <template #title>
      <div class="entity-header">
        <div class="entity-icon" @click="handleIconClick">
          <Icon :icon="getEntityIcon(expose.name, expose.data)" size="50" background="red" />

          <!-- <Icon
            :icon="{ name: mdiCeilingLightMultiple, color: lightOn ? '#ffc107' : '#9e9e9e' }"
            width="28"
            height="28" /> -->
        </div>
        <div class="entity-labels">
          <div class="entity-title">{{ expose.name }}</div>
          <div class="entity-value">{{ expose.data }}</div>
        </div>
      </div>
    </template>
    <template #content>
      <div>
        <!-- <Brightness :value="value" @update="updateValue" :min="0" :max="100" /> -->
      </div>
    </template>
  </Card>
</template>
