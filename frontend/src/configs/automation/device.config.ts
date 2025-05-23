import { AutomationActionStep, AutomationStepAction, AutomationTriggerAction } from '@/types/automation.type';
import { DeviceFilter, Expose, Device, ExposeType, ExposeCategory } from '@/types/device';
import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';

export function featureDevicesFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.access_mode != ExposeAccessModes.Read;
  };
}

export function stateDevicesFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return isStateExpose(expose);
  };
}

export function writableExposesDeviceFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return (
      isWritablePresetExpose(expose) ||
      isStateExpose(expose) ||
      isWritableEnumExpose(expose) ||
      isWritableMumericExpose(expose)
    );
  };
}

export function writableConfigPresetsExposesDeviceFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return (
      expose.access_mode == ExposeAccessModes.Write &&
      expose.category == ExposeCategories.Config &&
      expose.values != null
    );
  };
}

export const isStateExpose = (expose: Expose): boolean => {
  return (
    expose.type == ExposeTypes.Binary &&
    expose.access_mode != ExposeAccessModes.Read &&
    expose.category == ExposeCategories.Measurement
  );
};

export const isWritableEnumExpose = (expose: Expose): boolean => {
  return (
    expose.type == ExposeTypes.Enum &&
    expose.access_mode != ExposeAccessModes.Read &&
    expose.category == ExposeCategories.Measurement
  );
};

export const isWritableMumericExpose = (expose: Expose): boolean => {
  return (
    expose.type == ExposeTypes.Numeric &&
    expose.access_mode != ExposeAccessModes.Read &&
    expose.category == ExposeCategories.Measurement
  );
};

export const isWritablePresetExpose = (expose: Expose): boolean => {
  return isPresetExpose(expose) && expose.category == ExposeCategories.Measurement;
};

export const isPresetExpose = (expose: Expose): boolean => {
  if (!expose.values) {
    return false;
  }
  return (
    (expose.type == ExposeTypes.Enum || expose.type == ExposeTypes.Numeric) && Object.keys(expose.values).length > 0
  );
};

export function presetsDevicesFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return isPresetExpose(expose);
  };
}

export function devicesFilterById(id: string): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return device.id == id;
  };
}

export function devicesFilterByActionStep(
  automationId: string,
  action: AutomationStepAction,
  step: AutomationActionStep
): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    if (action.steps.length == 1) {
      step.id = action.id;
      return device.id == action.id;
    }

    return device.id == action.id || device.id == automationId;
  };
}

export function featureExposeFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.access_mode != ExposeAccessModes.Read;
  };
}

export function presetExposeFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return isPresetExpose(expose);
  };
}

export function exposeFilterByType(exposeType: ExposeType): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.type == exposeType;
  };
}

export function exposeFilterByTypeAndCategory(exposeType: ExposeType, category: ExposeCategory): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.type == exposeType && expose.category == category;
  };
}
export function allExposeFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return true;
  };
}
