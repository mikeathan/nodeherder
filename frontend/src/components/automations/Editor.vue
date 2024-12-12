<script setup lang="ts">
import { useRouter } from 'vue-router'
import DeviceAutomation from "./DeviceAutomation.vue"

const props = defineProps({
    id: String,
});

const router = useRouter()
function cancel(): void {
    router.push("/viewer")
}

</script>
<template>
    <h3>Automation editor</h3>
    <div>
        <DeviceAutomation :id="props.id" @cancel="cancel"></DeviceAutomation>
    </div>
</template>
