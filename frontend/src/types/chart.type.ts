export type ChartColor = {
  backgroundColor: string;
  borderColor: string;
};

export type AreaDataPoint = {
  x: string;
  y: number;
};

export type AreaChartEntry = {
  name: string;
  data: AreaDataPoint[];
};

export type TimelineDataPoint = {
  x: string;
  y: number[];
};

export type TimelineChartEntry = {
  name: string;
  data: TimelineDataPoint[];
};
