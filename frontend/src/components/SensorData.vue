<script setup>
const props = defineProps({
  device: Object
})
</script>

<template>
  <ol>{{ device.note }}</ol>
</template>