import { IconProps } from '@/types/icon.type';
import {
  mdiDotsVertical,
  mdiCrosshairsQuestion,
  mdiCancel,
  mdiCloudOffOutline,
  mdiAppleKeyboardCommand,
  mdiBroadcast,
} from '@mdi/js';

export type IconType = 'settings' | 'disabled' | 'offline' | 'command' | 'broadcast';

export function getIconForType(type: IconType): IconProps {
  switch (type) {
    case 'settings':
      return { name: mdiDotsVertical, color: 'grey' };

    case 'disabled':
      return { name: mdiCancel, color: 'white' };

    case 'offline':
      return { name: mdiCloudOffOutline, color: 'white' };

    case 'command':
      return { name: mdiAppleKeyboardCommand, color: 'white' };
    case 'broadcast':
      return { name: mdiBroadcast, color: 'white' };

    default:
      return { name: mdiCrosshairsQuestion, color: 'grey' };
  }
}
