export default {
  beforeMount(el: HTMLElement, binding: any) {
    const handler = (event: MouseEvent) => {
      if (!(el === event.target || el.contains(event.target as Node))) {
        binding.value(event);
      }
    };

    (el as any).__clickOutsideHandler__ = handler;

    document.addEventListener('mousedown', handler);
  },

  unmounted(el: HTMLElement) {
    document.removeEventListener('mousedown', (el as any).__clickOutsideHandler__);
    delete (el as any).__clickOutsideHandler__;
  },
};
