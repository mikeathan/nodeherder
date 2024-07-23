import { ChartColor } from '@/types/chart.type';


export const chartColors: ChartColor[] = buildColors(20);

var dynamicColors = function () {
  var r = Math.floor(Math.random() * 255);
  var g = Math.floor(Math.random() * 255);
  var b = Math.floor(Math.random() * 255);

  return {
    backgroundColor: 'rgb(' + r + ',' + g + ',' + b + ')',
    borderColor:
      'rgba(' + r + ',' + g + ',' + b + ',' + 0.5 + ')',
  };
};

function buildColors(n: number): ChartColor[] {
  let chartColors: ChartColor[] = [];
  for (let i = 0; i < n; i++) {
    chartColors[i] = dynamicColors();
  }
  return chartColors;
}
