import type { HelpTopic } from '@/domain/integrations/help';

/** Guidance for regional severe-weather warnings and provider coverage. */
export function alertsHelp(providerId: string): readonly HelpTopic[] {
  const coverage =
    providerId === 'met_office'
      ? 'Met Office Warnings covers the United Kingdom. Enter a UK town or postcode and choose the matching location.'
      : providerId === 'nws'
        ? 'National Weather Service covers the United States. Enter a US town or ZIP code; UK postcodes do not return results.'
        : 'Choose the provider for your region, then enter a town or postcode and choose the matching location.';
  return [
    {
      id: 'weather-alerts-setup',
      title: 'Weather alert setup',
      paragraphs: [
        coverage,
        'Warnings are the provider’s active warnings at the time of the check. “No active warnings” is a valid result, not a failure.',
        'Refresh checks the same saved location for newer warnings. Warning text is shown as supplied by the provider; follow official local guidance during an emergency.',
      ],
    },
  ];
}
