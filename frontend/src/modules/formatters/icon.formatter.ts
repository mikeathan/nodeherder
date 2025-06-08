import { IconProps } from '@/types/icon.type';
import { mdiDotsVertical, mdiCrosshairsQuestion } from '@mdi/js';

export type IconType = 'settings';
export function getIconForType(type: IconType): IconProps {
  switch (type) {
    case 'settings':
      return { name: mdiDotsVertical, color: 'grey' };

    default:
      return { name: mdiCrosshairsQuestion, color: 'grey' };
  }
}
