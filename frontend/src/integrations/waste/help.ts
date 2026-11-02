import type { HelpTopic } from '@/domain/integrations/help';

/** Provider-independent guidance for address-based collection schedules. */
export function wasteHelp(): readonly HelpTopic[] {
  return [
    {
      id: 'waste-collection-setup',
      title: 'Bin collection setup',
      paragraphs: [
        'Choose the provider for your area, enter the postcode, and select the exact address. The provider returns the collection schedule for that address.',
        'Collection dates are calendar dates. A missing date or an empty list does not prove a collection was cancelled. Refresh asks the provider for newer dates, and the saved schedule can also refresh automatically.',
      ],
    },
  ];
}
