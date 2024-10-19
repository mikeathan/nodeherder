import 'jest';
import { describe, expect, it } from '@jest/globals';
import { formatTimestamp as formatUnixTimestamp } from '../../utils/date.utils';

it('expect formatTimestamp to return correct formatted date', () => {
  //const timestamp = 1729335970554
  //1729336032729
  const timestamp = 1729335171422;
  const formattedDate = '2024-10-19T10:52:51.422Z';
  const date = formatUnixTimestamp(timestamp);

  expect(date).toBe(formattedDate);
});
