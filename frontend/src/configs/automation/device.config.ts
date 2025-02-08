import { AutomationActionStep, AutomationStepAction, AutomationTriggerAction } from '@/types/automation.type';
import { DeviceFilter, Expose, Device, ExposeType } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';

export function featureDevicesFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.properties != undefined;
  };
}

export function presetsDevicesFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.type == ExposeTypes.Enum || expose.presets != null;
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
    return expose.properties != undefined;
  };
}

export function presetExposeFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.type == ExposeTypes.Enum || expose.presets != null;
  };
}

export function exposeFilterByType(exposeType: ExposeType): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return expose.type == exposeType;
  };
}

export function allExposeFilter(): DeviceFilter {
  return (device: Device, expose: Expose): boolean => {
    return true;
  };
}
