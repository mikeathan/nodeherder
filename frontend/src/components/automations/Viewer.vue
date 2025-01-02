<script setup lang="ts">
  import { computed } from 'vue';
  import { RouterLink, useRouter } from 'vue-router';
  import { store } from '../../store/index';
  import { Automations } from '@/types/automation';
  import AutomationStatus from './schedule/AutomationStatus.vue';

  const router = useRouter();
  const automations = computed(() => {
    if (
      !store.getters['automations/initialized']() as Boolean
    ) {
      store.dispatch('ws/emit', {
        event: 'loadAutomations',
      });
    }
    return store.getters[
      'automations/listAll'
    ]() as Automations;
  });

  function onDeleteAutomationClick(id: string): void {
    // emit delete event
    store.dispatch('ws/emit', {
      event: 'deleteAutomation',
      message: {
        id: id,
      },
    });
  }

  const navigateToCreator = () => {
    router.push('/creator');
  };
</script>

<style scoped>
  .container {
    margin: 0 auto;
    padding: 20px;
  }

  .list {
    padding: 0;
    margin: 0;
  }

  .list-item {
    display: flex;
    flex-direction: column;
    padding: 16px;
    border-bottom: 1px solid #e0e0e0;
    gap: 16px;
  }

  .badge {
    display: flex;
    justify-content: center;
    align-items: center;
    color: white;
    font-weight: bold;
    width: 48px;
    height: 48px;
  }

  .details {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex-grow: 1;
  }

  .friendly-name {
    font-size: 18px;
    font-weight: 600;
  }

  .link {
    text-decoration: none;
  }

  .description {
    font-size: 14px;
    color: #4b5563;
    line-height: 1.5;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .actions {
    display: flex;
    gap: 16px;
    align-items: center;
  }

  /* Mobile: Stack vertically */
  @media (max-width: 768px) {
    .list-item {
      flex-direction: column;
    }

    .actions {
      justify-content: flex-start;
      margin-top: 8px;
    }
  }

  @media (min-width: 768px) {
    .list-item {
      flex-direction: row;
      align-items: center;
    }

    .actions {
      justify-content: flex-end;
    }

    .description {
      white-space: normal;
    }
  }
</style>
<template>
  <Card>
    <template #title>
      <h2>Automations</h2>
    </template>
    <template #content>
      <Divider type="solid" />
      <div class="container">
        <ul class="list">
          <template
            v-for="(automation, index) in automations"
            :key="automation.id">
            <li class="list-item">
              <div class="details">
                <div class="friendly-name">
                  <RouterLink
                    :to="`/editor/${automation.id}`"
                    class="link">
                    {{ automation.friendlyname }}
                  </RouterLink>
                </div>
                <div class="description">
                  {{ automation.description }}
                </div>
              </div>

              <!-- Actions Section -->
              <div class="actions">
                <AutomationStatus
                  :automation="automation" />
                <Button
                  icon="pi pi-trash"
                  variant="text"
                  rounded
                  @click="
                    onDeleteAutomationClick(automation.id)
                  " />
              </div>
            </li>
          </template>
        </ul>
      </div>

      <div class="pt-3"></div>
      <div class="col md:col-3 sm:col-6">
        <Button
          style="width: 60%"
          icon="pi pi-plus"
          label="Create automation"
          @click="navigateToCreator"
          size="small" />
      </div>
    </template>
  </Card>
</template>
