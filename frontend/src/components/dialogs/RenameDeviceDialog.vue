<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps({
    isOpen: Boolean,
    device: Object
});

const message = ref('This is a modal popup');
const emit = defineEmits(['close']);

const closeModal = () => {
    emit('close');
};
</script>
<style scoped>
.modal-mask {
    position: fixed;
    z-index: 9998;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    display: table;
    transition: opacity 0.3s ease;
}

.modal-wrapper {
    display: table-cell;
    vertical-align: middle;
}

.modal-container {
    width: 300px;
    margin: 0px auto;
    padding: 20px 30px;
    background-color: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.33);
    transition: all 0.3s ease;
    font-family: Helvetica, Arial, sans-serif;
}

.modal-header h3 {
    margin-top: 0;
    color: #42b983;
}

.modal-body {
    margin: 20px 0;
}

.modal-default-button {
    float: right;
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
    -webkit-transform: scale(1.1);
    transform: scale(1.1);
}
</style>
<template>
    https://codepen.io/immarina/pen/oNxXWKa

    <div class="modal-mask" v-if="isOpen" @click.self="closeModal">
        <div class="modal-wrapper">
            <div class="modal-container">

                <div class="modal-header">
                    <h3>Rename device</h3>
                    <small>{{ props.device.friendly_name }}</small>
                    <button @click.prevent="closeModal">X</button>
                </div>
            </div>
        </div>
    </div>

    <!-- <div v-if="isOpen" class="modal-mask" @click.self="closeModal">
        <div class="modal-wrapper">
            <div class="popup-header">
                <h3>Rename device</h3>
                <small>{{ props.device.friendly_name }}</small>
                <button @click.prevent="closeModal">X</button>
            </div>
            <article>
                <div class="popup-content-text">
                    This is a simple modal popup in Vue.js
                </div>
            </article>
        </div>
    </div> -->
</template>