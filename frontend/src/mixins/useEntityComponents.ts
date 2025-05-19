import { getExposeAttribute } from '@/contracts/device';
import { buildEventHandlers } from '@/contracts/events';
import { ControlDirection } from '@/types/controls.type';
import { Expose, ExposeType } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { EventActions } from '@/types/events.type';
import { h, defineAsyncComponent } from 'vue';

const FillSlider = defineAsyncComponent(() => import('../components/input/FillSlider.vue'));
const PickSlider = defineAsyncComponent(() => import('../components/input/PickSlider.vue'));

type ComponentResolver = (expose: Expose) => ReturnType<typeof defineAsyncComponent> | null;

const componentResolvers: Record<ExposeType, ComponentResolver> = {
  [ExposeTypes.Numeric]: (expose) => {
    const Component = expose.values ? PickSlider : FillSlider;
    return Component;
  },
};

type EntityInputProps = {
  direction?: ControlDirection;
  min?: number;
  max?: number;
  disabled?: boolean;
  unit?: string;
};

function buildExposeParams(expose: Expose, direction: ControlDirection): EntityInputProps {
  switch (expose.type) {
    case ExposeTypes.Numeric:
      const props = {
        min: getExposeAttribute(expose, 'min'),
        max: getExposeAttribute(expose, 'max'),
        direction: direction,
        unit: expose.unit,
        ticks:expose.values
        // disabled: expose.data == 0,
      };

      return { ...props };
  }
  return {} as EntityInputProps;
}

export const EntityInputComponents = (
  expose: Expose,
  direction: ControlDirection,
  events: EventActions
): ReturnType<typeof h> | null => {
  const props: EntityInputProps = buildExposeParams(expose, direction);
  const eventHandlers = buildEventHandlers(events);

  const Component = componentResolvers[expose.type](expose);
  return Component ? h(Component, { ...props, ...eventHandlers }) : null;
};
