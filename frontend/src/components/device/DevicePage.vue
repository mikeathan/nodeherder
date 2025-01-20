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

  const deviceExist = computed<boolean>(() => {
    return store.getters['hub/deviceExists'](props.id);
  });

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
      <div className="d-flex flex-row">
        <div class="align-self-center me-3">
          <RouterLink :to="`${previousPage}`">
            <Button icon="pi pi-arrow-left" variant="text" />
          </RouterLink>
        </div>
        <div class="h3 align-self-center">
          {{ device.friendly_name }}
        </div>
      </div>
    </template>
    <template #content>
      <div class="col-12 col-md-9">
        <Tabs value="0" class="flex flex-wrap gap-2">
          <TabList>
            <Tab
              v-for="tab in deviceTabComponents"
              :key="tab.title"
              :value="tab.value"
              class="flex-1 text-center p-2 md:flex-none">
              {{ tab.title }}
            </Tab>
          </TabList>
          <TabPanels>
            <TabPanel v-for="tab in deviceTabComponents" :key="tab.value" :value="tab.value">
              <component :is="tab.content" v-bind="{ id: props.id }"></component>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </template>
  </Card>
</template>
