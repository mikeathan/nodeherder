import { EventActions } from '@/types/events.type';

export function buildEventHandlers(events: EventActions): Record<string, Function> {
  const handlers: Record<string, Function> = {};

  for (const [eventName, handler] of Object.entries(events)) {
    const key = 'on' + eventName.charAt(0).toUpperCase() + eventName.slice(1);
    handlers[key] = handler;
  }
  return handlers;
}
