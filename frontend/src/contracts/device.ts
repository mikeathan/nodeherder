import { Expose } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";

export function getExposeAttribute(expose: Expose, name: string): any {
  return expose.attributes ? expose.attributes[name] : null;
}

export function getExposeProperty(expose: Expose, name: string): any {
  return expose.properties ? expose.properties[name] : null;
}

export function getExposeBinaryProperty(expose: Expose): boolean | null {
  if (expose.properties == null) {
    return null;
  }
  if (expose.data == expose.properties["on"]) {
    return true;
  }
  if (expose.data == expose.properties["off"]) {
    return false;
  }
  return null;
}
