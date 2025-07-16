import { computed } from 'vue';
import type { MenuBarItem } from '@/types/controls.type';

export function useMenuItems(items: MenuBarItem[], onItemSelected?: () => void) {
  const menuItems = computed(() =>
    items
      .filter((item) => !item.custom && !item.isLogo)
      .map((item) => ({
        label: item.label,
        icon: item.icon,
        disabled: item.disabled,
        command: () => {
          item.command?.();
          onItemSelected?.();
        },
        items: item.children?.map((child) => ({
          label: child.label,
          icon: child.icon,
          command: () => {
            child.command?.();
            onItemSelected?.();
          },
        })),
      }))
  );

  const logoItem = computed(() => items.find((item) => item.isLogo));

  return {
    menuItems,
    logoItem,
  };
}
