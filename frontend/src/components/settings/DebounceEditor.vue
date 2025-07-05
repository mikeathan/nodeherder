<script setup lang="ts">
  import { PropType, ref, watchEffect, watch } from 'vue';
  import Selection from '../input/Selection.vue';
  import TimeIntervalEditor from '../controls/TimeInterval.vue';
  import { TimeInterval } from '@/types/types.type';
  import { createTimeIntervalFromSeconds } from '@/contracts/settings';

  type DebounceMap = Record<string, TimeInterval>;

  const props = defineProps({
    value: { type: Object as PropType<DebounceMap>, required: true },
    disabled: { type: Boolean, default: false },
    selectableKeys: { type: Array as PropType<string[]>, default: () => [] },
    label: { type: String, default: 'Key' },
  });

  const emit = defineEmits<{
    (e: 'update', value: DebounceMap): void;
    (e: 'add-key', cb: (key: string) => void): void;
  }>();

  const selectedKey = ref<string | null>();
  const items = ref<DebounceMap>(props.value);
  const keyList = ref<string[]>([]);

  watchEffect(() => {
    keyList.value = Object.keys(items.value);
  });

  watch(
    () => props.value,
    () => {
      if (!props.value) return;
      selectedKey.value = firstKey();
    },
    { immediate: true }
  );

  function firstKey(): string | null {
    const keys = Object.keys(props.value);
    return keys.length > 0 ? keys[0] : null;
  }

  function removeSelectedKey() {
    if (!selectedKey.value) return;
    delete items.value[selectedKey.value];
    selectedKey.value = firstKey();
    emit('update', { ...items.value });
  }

  function addNewKey(key: string) {
    if (!key) return;
    items.value[key] = createTimeIntervalFromSeconds(1);
    selectedKey.value = key;
    emit('update', { ...items.value });
  }

  function updateDebounce(val: TimeInterval) {
    if (!selectedKey.value) return;
    items.value[selectedKey.value] = val;
    emit('update', { ...items.value });
  }

  function openAddDialog() {
    emit('add-key', (key: string) => addNewKey(key));
  }
</script>

<template>
  <div class="grid">
    <div class="col-12 sm:col-10 flex items-center">
      <Selection
        :label="label"
        :value="selectedKey"
        :items="keyList"
        @updated="(v: any) => { selectedKey = v }"
        :disabled="!keyList.length || disabled" />
      <div class="flex ml-2">
        <Button
          icon="pi pi-trash"
          variant="text"
          rounded
          size="small"
          :disabled="!selectedKey || disabled"
          @click="removeSelectedKey" />
        <Button icon="pi pi-plus" variant="text" rounded size="small" @click="openAddDialog" :disabled="disabled" />
      </div>
    </div>
  </div>
  <div v-if="selectedKey">
    <TimeIntervalEditor :value="items[selectedKey]" @update="updateDebounce" :disabled="disabled" />
  </div>
</template>
