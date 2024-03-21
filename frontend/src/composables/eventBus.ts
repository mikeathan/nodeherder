import { Events } from "@/types/events.type";
import mitt, { Emitter } from "mitt";
import { InjectionKey, onUnmounted, provide } from "vue";

export function useMittEvent<
  T extends Record<string, unknown>,
  K extends keyof T
>(mitt: Emitter<T>, key: K, fn: (e: T[K]) => any) {
  mitt.on(key, fn);
  onUnmounted(() => mitt.off(key, fn));
}

export type Handler<T = unknown> = (event: T) => void;
export type EventHandlers<T extends Record<string, unknown>> = {
  [K in keyof T]: (event: T[K]) => void;
};

export function useMittEvents<T extends Record<string, unknown>>(
  mitt: Emitter<T>,
  handlers: EventHandlers<T>
) {
  for (const key of Object.keys(handlers)) {
    mitt.on(key, handlers[key]);
  }
  function cleanup() {
    for (const key of Object.keys(handlers)) {
      mitt.off(key, handlers[key]);
    }
  }
  onUnmounted(cleanup);
  return cleanup;
}
// TESTING

type EventBus = {
  foo: string;
  bar: number;
};
export const eventBus = mitt<EventBus>();
export function useMyEvents(handlers: EventHandlers<EventBus>) {
  const keys = Object.keys(handlers) as Array<keyof EventBus>;
  for (const key of keys) {
    eventBus.on(key, handlers[key] as never);
  }
  const cleanup = () => {
    for (const key of keys) {
      eventBus.off(key, handlers[key] as never);
    }
  };
  onUnmounted(cleanup);
  return cleanup;
}

////
export const LocalEventBus: InjectionKey<Emitter<Events>> =
  Symbol("myEventBus");

// export const localBus = mitt<EventBus>();
// provide(LocalEventBus, localBus);

// https://dev.to/razi91/event-bus-with-vue-3-and-typescript-2a6l
// function wrapEventBus<T extends Record<string, unknown>>(
//   mitt: Emitter<T>
// ) {
//   return (ev: EventHandlers<T>) => useMittEvents(mitt, ev);
// }

// const useMyEvents = wrapEventBus(eventBus);
