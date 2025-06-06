export const getDeviceGroupId = (deviceId: string, expose: string): string => {
  return `${deviceId}-${expose}`;
};
