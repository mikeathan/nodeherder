import { goTo } from '@/router/navigation';
import { MenuBarItem } from '@/types/controls.type';
import { RouteName } from '@/types/router';
import { computed, h, ref } from 'vue';
import Logo from '@/components/controls/Logo.vue';

export function useSideNavigationItems() {
  const isPermitJoinActive = ref<boolean>(false);

  const startPermitJoinTimer = () => {
    if (!isPermitJoinActive.value) {
      isPermitJoinActive.value = true;
    }
  };
  const topNavigationItems = computed<MenuBarItem[]>(() => [
    {
      isLogo: true,
      template: () => h(Logo),
      command: () => {},
    },
  ]);
  const sideNavigationItems = computed<MenuBarItem[]>(() => [
    {
      label: 'groups',
      icon: 'pi pi-home',
      command: () => goTo(RouteName.GroupDashboard),
    },
    {
      label: 'devices',
      icon: 'pi pi-mobile',
      command: () => goTo(RouteName.Devices),
    },
    {
      to: '/devicelist',
      label: 'device list',
      icon: 'pi pi-list',
      command: () => goTo(RouteName.DeviceList),
    },
    {
      to: '/viewer',
      label: 'automations',
      icon: 'pi pi-objects-column',
      command: () => goTo(RouteName.Viewer),
    },
    {
      to: '/consoleviewer',
      label: 'console',
      icon: 'pi pi-code',
      command: () => goTo(RouteName.ConsoleViewer),
    },
    {
      to: '/settings',
      label: 'settings',
      icon: 'pi pi-cog',
      command: () => goTo(RouteName.Settings),
    },
    {
      label: 'permit Join',
      icon: 'pi pi-sitemap',
      get disabled() {
        return isPermitJoinActive.value;
      },
      command: () => startPermitJoinTimer(),
    },
  ]);
  return { sideNavigationItems, topNavigationItems, isPermitJoinActive };
}
