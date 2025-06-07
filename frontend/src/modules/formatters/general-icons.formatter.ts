import { IconProps } from '@/types/icon.type';
import { mdiDotsVertical } from '@mdi/js';

export function getIcon(type: any): IconProps {
  return { name: mdiDotsVertical, color: 'grey' };
}
