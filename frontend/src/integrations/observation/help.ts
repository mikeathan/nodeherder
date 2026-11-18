import type { HelpTopic } from '@/domain/integrations/help';

/**
 * Guidance for the local observation provider.
 *
 * It exists because this integration is the only one whose input is not a remote service. Every
 * shared help topic is about addresses, requests and credentials, so an operator reading them would
 * find nothing about the two things that actually go wrong here: the sensor must have Bluetooth
 * proxying compiled in, and the numbers reset.
 */
export function observationHelp(): readonly HelpTopic[] {
  return [
    {
      id: 'observation-sensor',
      title: 'Which sensor can do this',
      paragraphs: [
        'This integration reads the Bluetooth proxy on a sensor you have already enrolled. It does not need an address or a credential, because it reads what the hub is already receiving from that device.',
        'If the radar stays empty while the sensor is connected, check that its firmware includes Bluetooth proxying. A device without it still reports its own entities, so the device page looks healthy either way.',
      ],
    },
    {
      id: 'observation-numbers',
      title: 'What the numbers mean',
      paragraphs: [
        'Devices in range is the count of distinct wireless devices heard recently. Rotating addresses appear as separate entries, so a single phone can be heard under more than one identity; the radar groups them.',
        'Packets are totals since the hub started and reset when it restarts. They answer whether the radio is delivering at all, not how many devices are nearby. A high packet count with no devices means traffic is arriving but no identity could be resolved from it.',
      ],
    },
    {
      id: 'observation-off',
      title: 'Turning observation off',
      paragraphs: [
        'Turning Observing off stops the sensor being read, so nothing new is recorded. It does not remove the integration, its name or its place on the board, and turning it back on resumes immediately.',
        'On a sensor that shares its radio with another function, such as a presence sensor, turning this off gives that function the radio time back.',
        'Removing the integration deletes it and its Home card. It does not change the sensor or its enrolment.',
      ],
    },
  ];
}
