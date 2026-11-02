import type { HelpTopic } from '@/domain/integrations/help';

/** Guidance for pollutant and pollen readings. */
export function airQualityHelp(): readonly HelpTopic[] {
  return [
    {
      id: 'air-quality-setup',
      title: 'Air quality setup',
      paragraphs: [
        'Search for a town or postcode and choose the exact location. Results show the pollutants returned by the provider and, when available, pollen levels.',
        'Air-quality readings are provider estimates and can lag the current hour. Refresh updates the snapshot; use the checked time and freshness shown with the readings.',
      ],
    },
  ];
}
