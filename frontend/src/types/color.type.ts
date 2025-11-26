import { ValueOf } from './types.type';

export type ChartColor = {
  name: ColorType;
  color: ColorValue;
};

export type ColorType = keyof typeof ColorTypes;
export type ColorValue = ValueOf<typeof ColorTypes>;

export const ColorTypes = {
  Coral: '#FF4560',
  SkyBlue: '#69d2e7',
  SpringGreen: '#90ee7e',
  EmeraldGreen: '#00C853',
  Red: '#D7263D',
  Blue: '#2983FF',
  NavyBlue: '#2E294E',
  Green: '#4caf50',
  Lime: '#c7f464',
  Grey: '#C4BBAF',
  Purple: '#A300D6',
  Yellow: '#F9C80E',
  Pink: '#d4526e',
  Orange: '#F86624',
  LightBrown: '#A5978B',
  ElectricViolet: '#7D02EB',
  MidnightViolet: '#662E9B',
  Turquoise: '#4ecdc4',
  SlateGrey: '#546E7A',
};
