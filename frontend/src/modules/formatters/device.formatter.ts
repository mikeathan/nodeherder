export function getPowerSourceIcon(
  power_source: string,
  value: number
): string {
  if (power_source == '') {
    return '';
  }

  if (power_source.toLowerCase().includes('mains')) {
    return 'fa fa-plug';
  }

  let batteryClass: string = '';
  const battery: number = value;
  if (battery == undefined || battery >= 85) {
    batteryClass += ' fa-battery-full';
  } else if (battery >= 75) {
    batteryClass += ' fa-battery-three-quarters';
  } else if (battery >= 50) {
    batteryClass += ' fa-battery-half';
  } else if (battery >= 25) {
    batteryClass += ' fa-battery-quarter';
  } else if (battery >= 10) {
    batteryClass += ` fa-battery-empty animation-blinking`;
  } else {
    return `fa-battery-empty animation-blinking text-danger`;
  }

  if (!batteryClass) {
    batteryClass = 'fa-question';
  }
  return batteryClass;
}
