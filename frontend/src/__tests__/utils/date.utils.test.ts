import 'jest';
import { describe, expect, it } from '@jest/globals';
import { formatTimestamp as formatUnixTimestamp } from '../../utils/date.utils';

it('expect formatTimestamp to return correct formatted date', () => {
  const timestamp = 1729335171422;
  const date = formatUnixTimestamp(timestamp);

  // en-GB format is DD/MM/YYYY, HH:mm:ss.sss
  // Note: Depending on the environment, it might have a comma or not.
  // The function returns `${base}.${ms}` where base is from toLocaleString('en-GB', ...)
  expect(date).toContain('19/10/2024');
  expect(date).toContain('11:52:51.422');
});
