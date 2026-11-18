/**
 * Presentation for the observation census.
 *
 * The census is deliberately thin: it says what is reporting and how much, because the radar is the
 * live view. This presenter exists so the integration page renders that honestly instead of falling
 * through to "No information has been saved yet", which would describe a working integration as an
 * empty one.
 *
 * It declares no source kinds. A census describes the radar as a whole rather than exposing rows a
 * Home card can be built from; the dedicated observation Home tile reads the live snapshot instead.
 */
import { defineAsyncComponent } from 'vue';
import type {
  IntegrationPresenter,
  IntegrationResource,
  IntegrationSetupPresentation,
  IntegrationSourcePresentation,
  SourceIcon as SourceIconModel,
} from '../types';
import type { ConnectionInstance, ObservationCensus } from '@/types/integrations';
import type { IconName } from '@/components/ui/icons';
import { formatLocalDateTime } from '@/utils/date.utils';
import { observationHelp } from './help';
import { OBSERVATION_ROUTE } from '@/domain/integrations/observation';

const UiIcon = defineAsyncComponent(() => import('@/components/ui/UiIcon.vue'));

function icon(name: IconName): SourceIconModel {
  return { component: UiIcon, props: { name } };
}

/** Narrow a resource to the census, so nothing here trusts a contract version it did not check. */
function census(resource: IntegrationResource): ObservationCensus | undefined {
  return resource.contract_version === 'observation_census/v1' ? resource : undefined;
}

/**
 * Setup copy that is never shown.
 *
 * A local provider is added in one action, so no wizard ever renders. The values exist because the
 * shared registry resolves setup copy through the capability and throws when it is missing — a
 * presenter without this would break the provider page for every other integration.
 */
function unusedSetup(): IntegrationSetupPresentation {
  return {
    title: 'Wireless observation',
    sectionTitle: 'Source',
    defaultName: 'Enrolled sensors',
    searchLabel: 'Sensor',
    searchButton: 'Continue',
    choiceLabel: 'Sensor',
    choicePlaceholder: 'Choose a sensor',
    searchMaxLength: 0,
    autocomplete: 'off',
    notice: 'This integration is added in one step.',
  };
}

export const observationPresenter: IntegrationPresenter = {
  contractVersion: 'observation_census/v1',
  integrationType: 'wireless_observation',
  setup: unusedSetup,
  widget: (instance: ConnectionInstance) => {
    const resource = instance.resources.map(census).find(Boolean);
    if (!resource) {
      return {
        rows: [],
        remaining: 0,
        remainingLabel: '',
        status: instance.enabled ? 'Waiting for the first reading' : 'Disabled',
        checked: null,
        // Reached even with nothing to show: the radar is where an operator looks when a card is
        // empty, so it is the one destination that must never be missing.
        destination: { name: OBSERVATION_ROUTE, label: 'radar' },
      };
    }
    return {
      rows: [
        {
          id: 'identities',
          name: 'Devices in range',
          // Both numbers, because "20" and "20 of 25 shown" answer different questions, and an
          // operator comparing this card to the radar needs to see the second one.
          //
          // This is the row's value slot. The timestamp belongs in `checked`, which the status line
          // renders once; writing it here too printed the same time twice and left the device count
          // invisible, because the board reads `date` and ignores `value`.
          date:
            resource.hidden > 0
              ? `${resource.identities - resource.hidden} of ${resource.identities}`
              : `${resource.identities}`,
          colour: null,
          nextDateISO: null,
          icon: icon('signal'),
        },
      ],
      remaining: 0,
      remainingLabel: '',
      status: instance.enabled ? 'Observing' : 'Disabled',
      checked: formatLocalDateTime(resource.observed_at),
      icon: icon('signal'),
      destination: { name: OBSERVATION_ROUTE, label: 'radar' },
    };
  },
  sourceKinds: {},
  sources: (_instance: ConnectionInstance, _resource: IntegrationResource): IntegrationSourcePresentation[] => [],
  icon: (_kind: string, _colour: string | null) => icon('signal'),
  help: observationHelp,
  results: defineAsyncComponent(() => import('./Results.vue')),
};
