import { Events } from "@/types/events.type";
import { EventHandlers, LocalEventBus } from "./eventBus";
import { inject } from "vue";

export function useMyEvents(handlers: EventHandlers<Events>) {
  // get current context event bus
  const eventBus = inject(LocalEventBus);
  if (eventBus == null)
    // if not found
    throw new Error("No event bus found within this context");
  // it's fine, use the instance
  const keys = Object.keys(handlers) as Array<keyof Events>;
  for (const key of keys) {
    eventBus.on(key, handlers[key] as never);
  }
  // ... clean

  console.log("useMyEvents: TODO clean");
}
