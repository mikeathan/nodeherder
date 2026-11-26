import type { AppNotification } from '@/types/notification.type';

export interface NotificationsModuleState {
  /** Newest first. Bounded by the store action; persisted by the store plugin. */
  items: AppNotification[];
}
