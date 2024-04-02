import { DeviceFilter, Expose } from "@/types/device";

export function featureDevicesFilter(): DeviceFilter {
  return (expose: Expose): boolean => {
    return expose.properties != undefined;
  };
}
