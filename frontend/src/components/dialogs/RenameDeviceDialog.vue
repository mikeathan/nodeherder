<script setup lang="ts">
import { ref, watchEffect, onMounted, watch } from 'vue';
import { Modal } from 'bootstrap'

const props = defineProps<{
    friendlyName: string
    show: boolean
}>()

const emit = defineEmits(['update:name', 'close']);


function rename(event: Event) {
    emit('update:name', friendlyName.value);
    close()
}

const friendlyName = ref<string>("")
const closeRef = ref<HTMLButtonElement | null>(null);
const modalRef = ref<HTMLElement | null>(null)
const showDialog = ref<boolean>(false)
let modal: Modal

onMounted(() => {
    if (modalRef.value) {
        modal = new Modal(modalRef.value)
    }
})

// https://shzhangji.com/blog/2022/06/11/use-bootstrap-v5-in-vue3-project/

watchEffect(() => friendlyName.value = props.friendlyName);

watch(
    () => props.show,
    () => {
        showDialog.value = props.show
        if (showDialog.value) {
            modal.show()
        } else {
            modal.hide()
        }
    }
);

function close() {
    emit('close', false)
}
function isValid() {
    return friendlyName.value != '' && friendlyName.value != props.friendlyName
}
</script>

<template>
    <div class="modal fade" tabindex="-1" aria-hidden="true" ref="modalRef">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Rename device</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    <input type="text" class="form-control" v-model="friendlyName"
                        @input="e => friendlyName = (e.target as HTMLInputElement).value">
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal" @click="close">Close</button>
                    <button type="button" class="btn btn-primary" :disabled="isValid() == false"
                        @click="rename">Rename</button>
                </div>
            </div>
        </div>
    </div>
</template>