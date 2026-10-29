import type { HomeCard, HomeIntegrationCard } from '@/types/home.type';
import type { ConnectionInstance } from '@/types/integrations.type';
import type { IntegrationSourcePresentation } from '@/integrations/types';
import { integrationSources, integrationSourceTitle, integrationSourceIcon } from '@/integrations/registry';
import { homeSourceKey } from './home-layout';
import { exposeLabel } from './exposes';
/** Shared source presentation is supplied by the owning integration capability. */
export type HomeSourcePresentation = IntegrationSourcePresentation;
/** Index accepted capability sources without decoding their resources or owning transport. */
export function homeIntegrationSources(instances: readonly ConnectionInstance[]): Map<string, HomeSourcePresentation> {
  const sources = new Map<string, HomeSourcePresentation>();
  for (const instance of instances)
    for (const source of integrationSources(instance)) sources.set(homeSourceKey(source.reference), source);
  return sources;
}

/** Resolve a stable source reference, preserving an explicit placeholder if it disappears. Never
 * substitute another service based on dates, labels or current display order. */
export function homeIntegrationCard(
  card: HomeIntegrationCard,
  sources: ReadonlyMap<string, HomeSourcePresentation>
): HomeSourcePresentation {
  const accepted = sources.get(homeSourceKey(card));
  if (accepted) return accepted;
  return {
    reference: card,
    icon: integrationSourceIcon(card.kind, null),
    title: integrationSourceTitle(card.kind),
    value: 'Unavailable',
    origin: 'Integration result unavailable',
    status: 'The integration or its saved result is unavailable. Remove this card from Home if it is no longer needed.',
    checked: '',
    colour: null,
    warning: true,
    rows: [],
  };
}

/** Label a placement from its current source presentation; missing placements stay explicit. */
export function homeCardTitle(
  card: HomeCard | undefined,
  sources: ReadonlyMap<string, HomeSourcePresentation>
): string {
  if (!card) return 'Card unavailable';
  if (card.kind === 'device') return exposeLabel(card.expose);
  return homeIntegrationCard(card, sources).title;
}
