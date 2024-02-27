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

</script>

<template>
    <div class="modal fade" tabindex="-1" aria-hidden="true" ref="modalRef">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Action Selector</h5>
                </div>
                <div class="modal-body">
                </div>
                <div class="modal-footer">
                </div>
            </div>
        </div>
    </div>
</template>