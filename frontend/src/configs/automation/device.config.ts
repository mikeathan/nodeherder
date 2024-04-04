import { DeviceFilter, Expose, ExposeFilter, ExposeType } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";

export function featureDevicesFilter(): DeviceFilter {
  return (expose: Expose): boolean => {
    return expose.properties != undefined;
  };
}

export function presetsDevicesFilter(): DeviceFilter {
  return (expose: Expose): boolean => {
    return expose.type == ExposeTypes.Enum || expose.presets != null;
  };
}

export function featureExposeFilter(): ExposeFilter {
  return (expose: Expose): boolean => {
    return expose.properties != undefined;
  };
}

export function presetExposeFilter(): ExposeFilter {
  return (expose: Expose): boolean => {
    return expose.type == ExposeTypes.Enum || expose.presets != null;
  };
}

export function exposeFilterByType(exposeType:ExposeType): ExposeFilter {
  return (expose: Expose): boolean => {
    return expose.type == exposeType;
  };
}

