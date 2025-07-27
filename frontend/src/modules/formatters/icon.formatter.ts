import { IconProps } from '@/types/icon.type';
import { mdiDotsVertical, mdiCrosshairsQuestion, mdiCancel, mdiCloudOffOutline } from '@mdi/js';

export type IconType = 'settings' | 'disabled' | 'offline';

export function getIconForType(type: IconType): IconProps {
  switch (type) {
    case 'settings':
      return { name: mdiDotsVertical, color: 'grey' };

    case 'disabled':
      return { name: mdiCancel, color: 'white' };

    case 'offline':
      return { name: mdiCloudOffOutline, color: 'white' };

    default:
      return { name: mdiCrosshairsQuestion, color: 'grey' };
  }
}
