import { Expose, ExposeType } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { h, defineAsyncComponent } from 'vue';

type DialogKey = string;
type Map = { [key: DialogKey]: (props?: any) => ReturnType<typeof h> };

const StyledSlider = defineAsyncComponent(() => import('../components/input/StyledSlider.vue'));

TODO;
function getExposeParams(expose: Expose): string {
  switch (expose.type) {
    case ExposeTypes.Numeric:
      if (expose.values?.length > 0) {
        return 'slider-pick';
      }
      return 'slider-fill';
  }
  return '';
}

export const EntityInputComponents: Map = {
  numeric: (props) => h(StyledSlider, { ...props }),
};

// const tesmap = {
//     "temperature": "tap",
//     "brightness":"fill"
// }

//<component :is="EntityInputComponents['numeric']({ modelValue: 50, min: 0, max: 100 })" />
