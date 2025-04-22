<script setup lang="ts">
  import { PropType, ref, watch } from 'vue';
  import { IconProps } from '../../types/icon.type';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      default: {} as IconProps,
      required: true,
    },
    rotationAngle: {
      type: String,
      default: '0',
      required: false,
    },
    size: {
      type: String,
      default: '20px',
    },
   
    background: {
      type: String,
      default: '#ccc',
    },
  });

  const translateX = ref(0);
  const translateY = ref(0);

  // Watch for changes in the 'icon' prop
  watch(
    () => props.icon,
    (newIcon) => {
      if (newIcon && newIcon.name) {
        calculateCenterOffset(newIcon.name);
      } else {
        translateX.value = 0;
        translateY.value = 0;
      }
    },
    { immediate: true }
  );

  function calculateCenterOffset(pathData: string) {
    // **Simplified Center Estimation (This is the key part to adjust)**
    // This is a very basic attempt and might not work well for all icons.
    // A more robust solution would involve parsing the path data.

    let minX = Infinity;
    let minY = Infinity;
    let maxX = -Infinity;
    let maxY = -Infinity;

    // Basic splitting of path commands (very rudimentary)
    const commands = pathData.split(/(?=[MLCQSZ])/i);

    commands.forEach((command) => {
      const parts = command.trim().split(/[ ,]+/);
      const type = parts[0].toUpperCase();

      for (let i = 1; i < parts.length; i += 2) {
        const x = parseFloat(parts[i]);
        const y = parseFloat(parts[i + 1]);

        if (!isNaN(x) && !isNaN(y)) {
          minX = Math.min(minX, x);
          minY = Math.min(minY, y);
          maxX = Math.max(maxX, x);
          maxY = Math.max(maxY, y);
        }
      }
    });

    const pathCenterX = (minX + maxX) / 2;
    const pathCenterY = (minY + maxY) / 2;
    const svgCenter = 15;

    // Apply a correction factor - you might need to tweak these values
    const correctionX = 0;
    const correctionY = 0;

    translateX.value = svgCenter - pathCenterX + correctionX;
    translateY.value = svgCenter - pathCenterY + correctionY;
  }
</script>
<style scoped>
  /* Temporary */
  .align-middle {
    margin-right: 0.2em;
  }
</style>
<template>
   <svg
    :width="props.size"
    :height="props.size"
    viewBox="0 0 30 30" 
    :aria-label="props.icon.tooltip"
    class="align-middle"
  >
    <title>{{ props.icon.tooltip }}</title>

    <!-- Background circle with larger radius (e.g., 15) -->
    <circle
      cx="15"
      cy="15"
      r="15"  
      :fill="background"
    />

    <!-- Rotated icon path -->
    <g :transform="`rotate(${rotationAngle} 15 15)`">
      <path :d="props.icon.name" :fill="props.icon.color || 'black'" />
    </g>
  </svg>
</template>
