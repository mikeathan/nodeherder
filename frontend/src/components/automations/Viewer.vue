<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, useRouter } from 'vue-router';
import { store } from '../../store/index';
import {
    Automation,
    Automations,
} from '@/types/automation';
import AutomationStatus from './schedule/AutomationStatus.vue';

const router = useRouter();

const automations = computed(() => {
    if (
        !store.getters['automations/initialized']() as Boolean
    ) {
        store.dispatch('ws/emit', { event: 'loadAutomations' });
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

function getIndex(item: Automation): number {
    return automations.value.indexOf(item);
}

const navigateToCreator = () => {
    router.push('/creator');
};
</script>

<template>

    <Card>
        <template #title>
            <h2>Automations</h2>
        </template>
        <template #content>
            <Divider type="solid" />
            <ul class="list-none p-0 m-0">
                <template v-for="(automation, index) in automations">
                    <li class="p-3 border-b-1  flex flex-wrap md:flex-nowrap items-center">

                        <!-- Index Badge -->
                        <div
                            class="flex-shrink-0 flex justify-center items-center text-white font-bold bg-blue-500 rounded-full me-3 mt-2 mb-2 md:mb-0">
                            <Badge severity="secondary" :value="index + 1" size="small"></Badge>
                        </div>

                        <!-- Automation Details -->
                        <div class="flex-1 mb-2 md:mb-0">
                            <!-- Friendly Name -->
                            <div class="text-lg font-medium">
                                <RouterLink :to="`/editor/${automation.id}`" class="text-blue-600 hover:underline">
                                    {{ automation.friendlyname }}
                                </RouterLink>
                            </div>
                            <!-- Description -->
                            <div class="text-sm text-gray-600">
                                {{ automation.description }}
                            </div>
                        </div>

                        <!-- Actions Section -->
                        <div class="flex-shrink-0 flex items-center space-x-3 w-full  mt-2 md:w-auto lg:w-30">
                            <!-- Status -->
                            <div>
                                <AutomationStatus :automation="automation" />
                            </div>
                            <div>
                                <!-- Delete Button -->
                                <Button icon="pi pi-trash" variant="text" rounded class="text-red-500"
                                    @click="onDeleteAutomationClick(automation.id)" />
                            </div>
                        </div>

                    </li>
                    <Divider type="solid" />
                </template>
            </ul>
            <div class="pt-3"></div>
            <div class="col md:col-3 sm:col-6">
                <Button style="width: 99%" icon="pi pi-plus" label="Create automation" @click="navigateToCreator"
                    size="small" />
            </div>
        </template>
    </Card>
</template>
