import { ref } from 'vue';
import { useChartAppearance } from '@/composables/history/useChartAppearance';

const tokens: Record<string, string> = {
  '--nh-kind-power': '#111111',
  '--nh-kind-temp': '#222222',
  '--nh-kind-contact': '#333333',
  '--nh-accent': '#444444',
};

beforeAll(() => {
  (global.window as unknown as { matchMedia: unknown }).matchMedia = jest.fn().mockReturnValue({
    matches: false,
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
  });
  (global as unknown as { getComputedStyle: unknown }).getComputedStyle = jest.fn(() => ({
    fontFamily: 'sans-serif',
    getPropertyValue: (name: string) => tokens[name] ?? '',
  }));
  const root = global.document.documentElement as unknown as {
    setAttribute: unknown;
    style: { setProperty: unknown; removeProperty: unknown };
  };
  root.setAttribute = jest.fn();
  root.style.setProperty = jest.fn();
  root.style.removeProperty = jest.fn();
});

test('provides the three statistics series colours from one owner', () => {
  const appearance = useChartAppearance();
  expect(appearance.value.series).toEqual({
    power: '#111111',
    temp: '#222222',
    contact: '#333333',
  });
});

test('reads compact through a getter so a viewport change is not frozen at mount', () => {
  const compact = ref(false);
  const appearance = useChartAppearance(undefined, () => compact.value);
  expect(appearance.value.compact).toBe(false);
  compact.value = true;
  // A one-shot boolean would still report false here, leaving axis/animation config stale.
  expect(appearance.value.compact).toBe(true);
});
