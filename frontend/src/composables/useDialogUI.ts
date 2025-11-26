import { ref, computed, onMounted, onUnmounted } from 'vue';


export function useDialogUI(onClose: () => void) {
  const isMobile = ref(false);

  const checkMobile = () => {
    if (typeof window !== 'undefined') {
      isMobile.value = window.innerWidth <= 640;
    } else {
      isMobile.value = false;
    }
  };

  const handleEscape = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      onClose();
    }
  };

  onMounted(() => {
    checkMobile();
    window.addEventListener('resize', checkMobile);
    window.addEventListener('keydown', handleEscape);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', checkMobile);
    window.removeEventListener('keydown', handleEscape);
  });

  const dialogStyle = computed(() => {
    if (isMobile.value) {
      return {
        width: '100vw',
        height: '100dvh',
        maxHeight: '100dvh',
        margin: '0',
        transform: 'none',
        borderRadius: '0',
        zIndex: '9999',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      };
    }
    return {
      width: 'auto',
      minWidth: '460px',
      maxWidth: '90vw',
      height: 'auto',
      maxHeight: '90vh',
      borderRadius: '1rem',
      overflow: 'auto',
      display: 'flex',
      flexDirection: 'column',
    };
  });

  return {
    isMobile,
    dialogStyle,
  };
}
