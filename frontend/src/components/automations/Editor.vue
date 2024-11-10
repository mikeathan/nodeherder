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
    <div class="container-fluid p-0 h-100">
        <h3>Automation editor</h3>
        <div>
            <DeviceAutomation :id="props.id" @cancel="cancel"></DeviceAutomation>
        </div>
    </div>
</template>
