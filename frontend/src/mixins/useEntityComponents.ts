import { getExposeAttribute } from '@/contracts/device';
import { buildEventHandlers } from '@/contracts/events';
import { ControlDirection, StyledSliderInputType } from '@/types/controls.type';
import { Expose } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { EventActions } from '@/types/events.type';
import { h, defineAsyncComponent } from 'vue';

const StyledSlider = defineAsyncComponent(() => import('../components/input/StyledSlider.vue'));

type EntityInputTypes = StyledSliderInputType;
type EntityInputProps = {
  type: EntityInputTypes;
  direction?: ControlDirection;
  min?: number;
  max?: number;
  disabled?: boolean;
};

function buildExposeParams(expose: Expose): EntityInputProps {
  switch (expose.type) {
    case ExposeTypes.Numeric:
      const props = {
        min: getExposeAttribute(expose, 'min'),
        max: getExposeAttribute(expose, 'max'),
        direction: 'vertical' as ControlDirection,
        // disabled: expose.data == 0,
      };

      if (expose.values) {
        return { ...props, type: 'Pick' };
      }
      return { ...props, type: 'Fill' };
  }
  return {} as EntityInputProps;
}

export const EntityInputComponents = (expose: Expose, events: EventActions): ReturnType<typeof h> | null => {
  const props: EntityInputProps = buildExposeParams(expose);
  const eventHandlers = buildEventHandlers(events);

  if (expose.type == ExposeTypes.Numeric) {
    return h(StyledSlider, { ...props, ...eventHandlers });
  }
  return null;
};
