<script setup lang="ts">
import { store } from "../../store/index";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import LastSeen from "../device/LastSeen.vue";
import PowerSource from "../device/PowerSource.vue";
import ConnectionType from "../device/ConnectionType.vue";
import RenameDeviceDialog from "../dialogs/RenameDeviceDialog.vue";
import { Device } from "@/types/device";

const props = defineProps({
    id: String,
});

</script>
<template>
    <div>
        EXPOSE Page
    </div>
</template>
