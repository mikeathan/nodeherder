<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { store } from '../../store/index';
  import { useRouter } from 'vue-router';
  import { deviceTabComponents } from '../../mixins/useTabComponents';
  import { Device } from '@/types/device';

  const props = defineProps({
    id: {
      type: String,
      required: true,
    },
  });

  const activeTab = ref('0');

  const device = computed<Device>(() => {
    return store.getters['hub/findDevice'](props.id);
  });

  const previousPage = computed(() => {
    const back = useRouter().options.history.state.back;
    if (back != undefined) {
      return back;
    }

    return useRouter().push('/');
  });
</script>
<template>
  <Card>
    <template #title>
      <div class="flex flex-row">
        <div class="align-self-center me-3">
          <RouterLink :to="`${previousPage}`">
            <Button icon="pi pi-arrow-left" variant="text" />
          </RouterLink>
        </div>
        <div class="text-3xl align-self-center">
          {{ device?.friendly_name || 'friendly_Name not found' }}
        </div>
      </div>
    </template>
    <template #content>
      <div class="col-12 md:col-12">
        <Tabs v-model:value="activeTab">
          <TabList>
            <Tab
              v-for="tab in deviceTabComponents"
              :key="tab.title"
              :value="tab.value">
              {{ tab.title }}
            </Tab>
          </TabList>
          <TabPanels>
            <TabPanel v-for="tab in deviceTabComponents" :key="tab.value" :value="tab.value">
              <component v-if="activeTab === tab.value" :is="tab.content" v-bind="{ id: props.id }"></component>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </template>
  </Card>
</template>
