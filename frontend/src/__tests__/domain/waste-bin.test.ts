import { wasteBinColours } from '@/domain/waste-bin';

test('recognised bin colours have distinct illustrations without provider-specific configuration', () => {
  const colours = ['black', 'blue', 'brown', 'green', 'red', 'yellow', 'orange', 'purple', 'grey'];
  expect(new Set(colours.map((colour) => wasteBinColours(colour).body)).size).toBe(colours.length);
  expect(wasteBinColours(' BLUE ')).toEqual(wasteBinColours('blue'));
  expect(wasteBinColours('gray')).toEqual(wasteBinColours('grey'));
});

test.each([null, '', 'not specified', '#ff0000', 'url(https://example.com)', '__proto__'])(
  'unknown %s stays neutral',
  (colour) => {
    expect(wasteBinColours(colour)).toEqual(wasteBinColours('grey'));
  }
);
