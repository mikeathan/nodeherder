import {
  activeAlertCountLabel,
  alertLevelLabel,
  alertSeverityLabel,
  alertSeverityTone,
} from '@/domain/integrations/weather-alerts';

test('present normalized severity and provider level labels', () => {
  expect(alertSeverityLabel('severe')).toBe('Severe');
  expect(alertSeverityLabel('unexpected')).toBe('Unknown');
  expect(alertLevelLabel('amber')).toBe('Amber');
  expect(alertSeverityTone('extreme')).toBe('severe');
  expect(alertSeverityTone('minor')).toBe('minor');
});

test('label active warning counts', () => {
  expect(activeAlertCountLabel(0)).toBe('No active warnings');
  expect(activeAlertCountLabel(1)).toBe('1 active warning');
  expect(activeAlertCountLabel(3)).toBe('3 active warnings');
});
