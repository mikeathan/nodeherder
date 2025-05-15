import { getExposeAttribute } from '@/contracts/device';
import { ControlDirection, StyledSliderInputType } from '@/types/controls.type';
import { Expose } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { h, defineAsyncComponent } from 'vue';

type DialogKey = string;
type Map = { [key: DialogKey]: (props?: EntityInputProps) => ReturnType<typeof h> };

const StyledSlider = defineAsyncComponent(() => import('../components/input/StyledSlider.vue'));

type EntityInputTypes = StyledSliderInputType;
type EntityInputProps = {
  type: EntityInputTypes;
  direction?: ControlDirection;
  min?: number;
  max?: number;
};

function getExposeParams(expose: Expose): EntityInputProps {
  switch (expose.type) {
    case ExposeTypes.Numeric:
      const props = {
        min: getExposeAttribute(expose, 'min'),
        max: getExposeAttribute(expose, 'max'),
        direction: 'vertical' as ControlDirection,
      };

      if (!expose.values) {
        return { ...props, type: 'Pick' };
      }
      return { ...props, type: 'Fill' };
  }
  return {} as EntityInputProps;
}

export const EntityInputComponents = (expose: Expose): ReturnType<typeof h> | null => {
  const props: EntityInputProps = getExposeParams(expose);
  if (expose.type == ExposeTypes.Numeric) {
    return h(StyledSlider, props);
  }
  return null;
};

//<component :is="EntityInputComponents['numeric']({ modelValue: 50, min: 0, max: 100 })" />
