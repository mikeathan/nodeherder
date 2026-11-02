import type { HelpTopic } from '@/domain/integrations/help';

/** Guidance for the paginated nationwide fuel-price sync and its saved area. */
export function fuelHelp(): readonly HelpTopic[] {
  return [
    {
      id: 'fuel-prices-setup',
      title: 'Fuel prices setup',
      paragraphs: [
        'Enter a UK town or postcode and choose the exact search area. Results list the nearest forecourts within the provider radius, with the grades and prices each station returns.',
        'The provider publishes nationwide data in pages, so the first collection can take a few minutes. Later refreshes use the saved schedule; each price also shows when that grade was last updated.',
        'Fuel Finder requires OAuth client credentials. Save them in Setup before searching; they are stored privately and sent only to the provider’s token endpoint.',
      ],
    },
  ];
}
