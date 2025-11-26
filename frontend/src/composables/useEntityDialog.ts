import { ref, computed, watchEffect } from 'vue';
import { store } from '@/store/index';
import { Device, Expose } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { getExposeBinaryProperty, getExposes, toggleExposeBinaryProperty } from '@/contracts/device';
import {
  writableExposesDeviceFilter,
  writableConfigPresetsExposesDeviceFilter,
} from '@/configs/automation/device.config';
import { EntityInputComponents } from '@/mixins/useEntityComponents';

/**
 * Composable for managing entity dialog state and logic
 */
export function useEntityDialog(deviceId: string, exposeName: string) {
  const selectedControlExpose = ref<Expose | null>(null);

  const device = computed(() => {
    return store.getters['hub/findDevice'](deviceId) as Device | null;
  });

  const expose = computed(() => {
    if (!device.value) return {} as Expose;
    return device.value.exposes[exposeName] as Expose;
  });

  const lastSeen = computed(() => {
    return device.value ? device.value.last_seen : '';
  });

  const configExposes = computed(() => {
    if (!device.value) return [];

    const exposeNameList = getExposes(device.value, writableConfigPresetsExposesDeviceFilter());
    return exposeNameList.map((name) => device.value!.exposes[name] as Expose);
  });

  const controlExposes = computed(() => {
    if (!device.value) return [];

    const exposeNameList = getExposes(device.value, writableExposesDeviceFilter());
    return exposeNameList.map((name) => device.value!.exposes[name] as Expose);
  });

  const selectedComponent = computed(() => {
    if (!selectedControlExpose.value) return null;

    return (
      EntityInputComponents(selectedControlExpose.value, 'vertical', {
        update: (e: any) => updateValue(selectedControlExpose.value!.name, e),
      }) ?? null
    );
  });

  // Pre-select expose control using priority order
  watchEffect(() => {
    if (!selectedControlExpose.value) {
      const selected = controlExposes.value.reduce<Expose | null>((acc, expose) => {
        if (acc) return acc;

        if (expose.type === ExposeTypes.Numeric) {
          if (expose.values === null) {
            return expose; // Highest priority
          }
          return acc ?? expose; // Lower-priority fallback
        }

        return acc;
      }, null);

      selectedControlExpose.value = selected;
    }
  });

  function isEnabled() {
    if (device.value?.availability === 'offline') {
      return false;
    }

    if (expose.value.type === ExposeTypes.Numeric && expose.value.data === 0) {
      return false;
    }

    if (controlExposes.value) {
      return controlExposes.value.some((exp) => {
        if (exp.type === ExposeTypes.Binary) {
          return getExposeBinaryProperty(exp);
        }
      });
    }
    return true;
  }

  function handleClick(expose: Expose) {
    if (device.value?.availability === 'offline') {
      return;
    }
    if (expose.type === ExposeTypes.Binary) {
      const value = toggleExposeBinaryProperty(expose);
      updateValue(expose.name, value);
    } else {
      selectedControlExpose.value = expose;
    }
  }

  function updateValue(exposeName: string, newValue: any): void {
    const msg = {
      id: deviceId,
      name: exposeName,
      value: newValue,
    };

    store.dispatch('hub/setDeviceValue', msg);
  }

  function buildMenuItems(expose: Expose) {
    if (!expose.values) {
      return [];
    }
    return Object.values(expose.values).map((value: any) => {
      return {
        label: value,
        value: value,
        command: () => {
          updateValue(expose.name, value);
        },
      };
    });
  }

  return {
    device,
    expose,
    lastSeen,
    configExposes,
    controlExposes,
    selectedControlExpose,
    selectedComponent,
    isEnabled,
    handleClick,
    updateValue,
    buildMenuItems,
  };
}
