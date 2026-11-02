import type { HelpTopic } from '@/domain/integrations/help';

/** Guidance for location-based current conditions and forecasts. */
export function weatherHelp(providerId: string): readonly HelpTopic[] {
  const coverage =
    providerId === 'met_office'
      ? 'Met Office weather covers the United Kingdom. Use a UK town or postcode.'
      : 'Location search and forecast coverage depend on the selected provider. Choose the result that matches your exact area.';
  return [
    {
      id: 'weather-setup',
      title: 'Weather setup',
      paragraphs: [
        `${coverage} Weather results show current conditions and the forecast returned by that provider.`,
        'Refresh updates the forecast. Provider update frequency and forecast range vary, so use the checked time and freshness shown with the result rather than assuming live data.',
      ],
    },
  ];
}
