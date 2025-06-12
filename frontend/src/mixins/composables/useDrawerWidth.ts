import { watch, computed, Ref } from 'vue';



TODO 



export function useDrawerWidth(isMobile: Ref<boolean>, isExpanded: Ref<boolean>, emitWidth: (w: number) => void) {
  const drawerWidth = computed(() => (isMobile.value ? 0 : isExpanded.value ? 250 : 60));

  const emitDrawerWidth = () => {
    emitWidth(drawerWidth.value);
  };

  watch([isExpanded, isMobile], emitDrawerWidth, { immediate: true });

  return {
    drawerWidth,
    emitDrawerWidth,
  };
}
