import { defineAsyncComponent } from 'vue';

type TabContent = {
  title: string;
  value: string;
  content: any;
};

export const deviceTabComponents: TabContent[] = [
  {
    title: 'About',
    value: '0',
    content: defineAsyncComponent(
      () => import('../components/device/DeviceAbout.vue')
    ),
  },
  {
    title: 'Exposes',
    value: '1',
    content: defineAsyncComponent(
      () => import('../components/device/DeviceExposes.vue')
    ),
  },
  {
    title: 'Settings',
    value: '2',
    content: defineAsyncComponent(
      () =>
        import('../components/device/DeviceSettings.vue')
    ),
  },
  {
    title: 'Metrics',
    value: '3',
    content: defineAsyncComponent(
      () => import('../components/device/DeviceMetrics.vue')
    ),
  },
];

export const settingsTabComponents: TabContent[] = [
  {
    title: 'Device Defaults',
    value: '0',
    content: defineAsyncComponent(
      () => import('../components/hub/settings/DeviceDefaultSettings.vue')
    ),
  },
  {
    title: 'History',
    value: '1',
    content: defineAsyncComponent(
      () => import('../components/hub/settings/HistorySettings.vue')
    ),
  },
  {
    title: 'Logger',
    value: '2',
    content: defineAsyncComponent(
      () => import('../components/hub/settings/LoggerSettings.vue')
    ),
  },
  {
    title: 'MCP Server',
    value: '3',
    content: defineAsyncComponent(
      () => import('../components/hub/settings/MCPSettings.vue')
    ),
  },
  {
    title: 'Assistant',
    value: '4',
    content: defineAsyncComponent(
      () => import('../components/hub/settings/AssistantSettings.vue')
    ),
  },
];
