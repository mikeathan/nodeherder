import { StyledSliderInputType } from '@/types/controls.type';
import { Expose } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { h, defineAsyncComponent } from 'vue';

type DialogKey = string;
type Map = { [key: DialogKey]: (props?: EntityInputProps) => ReturnType<typeof h> };

const StyledSlider = defineAsyncComponent(() => import('../components/input/StyledSlider.vue'));

type EntityInputTypes = StyledSliderInputType;
type EntityInputProps = {
  type: EntityInputTypes;
};

function getExposeParams(expose: Expose): EntityInputProps {
  switch (expose.type) {
    case ExposeTypes.Numeric:
      if (expose.values?.length > 0) {
        return { type: 'Pick' };
      }
      return { type: 'Fill' };
  }
  return {} as EntityInputProps;
}

export const EntityInputComponents = (expose: Expose): Map => {
  const props: EntityInputProps = getExposeParams(expose);
  return {
    numeric: () => h(StyledSlider, props),
  };
};

//<component :is="EntityInputComponents['numeric']({ modelValue: 50, min: 0, max: 100 })" />
