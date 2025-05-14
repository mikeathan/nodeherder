import { StyledSliderInputType } from '@/types/controls.type';
import { Expose, ExposeType } from '@/types/device';
import { ExposeTypes } from '@/types/device.type';
import { Style } from 'primevue';
import { h, defineAsyncComponent } from 'vue';

type DialogKey = string;
type Map = { [key: DialogKey]: (props?: any) => ReturnType<typeof h> };

const StyledSlider = defineAsyncComponent(() => import('../components/input/StyledSlider.vue'));



function getExposeParams(expose: Expose): StyledSliderInputType |null{
  switch (expose.type) {
    case ExposeTypes.Numeric:
      if (expose.values?.length > 0) {
        return 'Pick';
      }
      return 'Fill';
  }
  return null;
}

// export const EntityInputComponents: Map = {
//   numeric: (props) => h(StyledSlider, { ...props }),
// };
export const EntityInputComponents = (expose: Expose): Map => {
  const props: StyledSliderInputType| null = getExposeParams(expose);
  return {
    numeric: () => h(StyledSlider, { ...props }),
  };
};

//<component :is="EntityInputComponents['numeric']({ modelValue: 50, min: 0, max: 100 })" />
