import { EventActions, Events, OpenPanelEvent } from "@/types/events.type";
import mitt, { Emitter } from "mitt";
import { InjectionKey, onUnmounted, provide } from "vue";

export type Handler<T = unknown> = (event: T) => void;
export type EventHandlers<T extends Record<string, unknown>> = {
  [K in keyof T]: (event: T[K]) => void;
};

const automationEventBus = mitt<Events>();

export function emitClosePanel(name: string) {
  automationEventBus.emit("closePanel", name);
}

export function emitOpenPanel(event: OpenPanelEvent) {
  automationEventBus.emit("openPanel", event);
}

export function useAutomationEvents(handlers: EventHandlers<Events>) {
  const keys = Object.keys(handlers) as Array<keyof Events>;
  for (const key of keys) {
    automationEventBus.on(key, handlers[key] as never);
  }

  const cleanup = () => {
    for (const key of keys) {
      automationEventBus.off(key, handlers[key] as never);
    }
  };
  onUnmounted(cleanup);
  return cleanup;
}

export function createActionOpenPanelEvent(
  name: string,
  args: any,
  events: EventActions
): OpenPanelEvent {
  return { name: name, args: args, events: events };
}

// https://dev.to/razi91/event-bus-with-vue-3-and-typescript-2a6l
