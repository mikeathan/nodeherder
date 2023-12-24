<script setup>
import { useRouter } from 'vue-router'
import DeviceAutomation from "./DeviceAutomation"


const props = defineProps({
    id: String,
});

const router = useRouter()


function cancel() {
    router.push("/viewer")
}
</script>
<template>
    <div class="container-fluid p-0 h-100">

        <h3>Edit Automation</h3>
        <div>
            <DeviceAutomation :id="props.id" @cancel="cancel"></DeviceAutomation>
        </div>
    </div>
</template>
