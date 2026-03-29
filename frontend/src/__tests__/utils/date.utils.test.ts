import 'jest';
import { describe, expect, it } from '@jest/globals';
import { formatTimestamp as formatUnixTimestamp } from '../../utils/date.utils';

it('expect formatTimestamp to return correct formatted date', () => {
  const timestamp = 1729335171422;
  const date = formatUnixTimestamp(timestamp);

  // Create a Date object to calculate expected values based on local timezone
  const testDate = new Date(timestamp);
  const expectedDay = String(testDate.getDate()).padStart(2, '0');
  const expectedMonth = String(testDate.getMonth() + 1).padStart(2, '0');
  const expectedYear = testDate.getFullYear();
  const expectedHours = String(testDate.getHours()).padStart(2, '0');
  const expectedMinutes = String(testDate.getMinutes()).padStart(2, '0');
  const expectedSeconds = String(testDate.getSeconds()).padStart(2, '0');
  const expectedMs = String(testDate.getMilliseconds()).padStart(3, '0');
  
  // The function uses toLocaleString('en-GB') which formats as DD/MM/YYYY, HH:mm:ss
  // We need to check both with and without comma since toLocaleString behavior varies
  const expectedDatePart = `${expectedDay}/${expectedMonth}/${expectedYear}`;
  const expectedTimePart = `${expectedHours}:${expectedMinutes}:${expectedSeconds}.${expectedMs}`;
  
  // Check that the date contains both parts
  expect(date).toContain(expectedDatePart);
  expect(date).toContain(expectedTimePart);
  
  // Also verify the format is correct (DD/MM/YYYY, HH:mm:ss.sss)
  // The comma might or might not be present depending on environment
  const dateRegex = new RegExp(`^${expectedDay}/${expectedMonth}/${expectedYear},?\\s*${expectedHours}:${expectedMinutes}:${expectedSeconds}\\.${expectedMs}$`);
  expect(date).toMatch(dateRegex);
});
