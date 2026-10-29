import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations.type';
import type {
  IntegrationPresenter,
  IntegrationResource,
  IntegrationSourcePresentation,
  IntegrationSourceReference,
  SourceIcon,
} from './types';
import { wastePresenter } from './waste/presentation';
import { weatherPresenter } from './weather/presentation';

// Composition is the only layer that names the installed presentation adapters.
const presenters: readonly IntegrationPresenter[] = [wastePresenter, weatherPresenter];

/** Resolve the presenter registered for a validated resource; absent capabilities fail visibly. */
export function resourcePresenter(resource: IntegrationResource): IntegrationPresenter | undefined {
  return presenters.find((presenter) => presenter.contractVersion === resource.contract_version);
}
/** Collect capability-owned Home sources without decoding payloads in the caller. */
export function integrationSources(instance: ConnectionInstance): IntegrationSourcePresentation[] {
  return instance.resources.flatMap((resource) => {
    const presenter = resourcePresenter(resource);
    if (!presenter) return [];
    return presenter
      .sources(instance, resource)
      .map((source) => ({ ...source, icon: source.icon ?? presenter.icon(source.reference.kind, source.colour) }));
  });
}
/** Resolve the label of a persisted source, retaining a generic fallback for removed capabilities. */
export function integrationSourceTitle(kind: string): string {
  return presenters.find((presenter) => kind in presenter.sourceKinds)?.sourceKinds[kind].title ?? 'Integration result';
}
/** Resolve capability-owned decorative icons without domain branches in generic cards. */
export function integrationSourceIcon(kind: string, colour: string | null): SourceIcon | undefined {
  return presenters.find((presenter) => kind in presenter.sourceKinds)?.icon(kind, colour);
}
/** Parse opaque reference identity according to the installed capability's source declaration. */
export function parseIntegrationReference(raw: Record<string, unknown>): IntegrationSourceReference {
  const kind = raw.kind;
  if (typeof kind !== 'string') throw new Error('Invalid integration reference');
  const declaration = presenters.find((presenter) => kind in presenter.sourceKinds)?.sourceKinds[kind];
  if (!declaration) throw new Error('Unsupported integration source');
  for (const key of ['instance_id', 'resource_id'])
    if (typeof raw[key] !== 'string' || !raw[key]) throw new Error('Missing integration reference');
  if (declaration.requiresService) {
    if (typeof raw.service_id !== 'string' || !raw.service_id) throw new Error('Missing integration sub-source');
  } else if (raw.service_id !== undefined && raw.service_id !== '')
    throw new Error('Unexpected integration sub-source');
  const reference = {
    kind,
    instance_id: raw.instance_id,
    resource_id: raw.resource_id,
    ...(declaration.requiresService ? { service_id: raw.service_id } : {}),
  };
  return reference as IntegrationSourceReference;
}

/** Resolve a connection's capability presenter; unknown categories retain an explicit empty state. */
export function connectionPresentation(instance: ConnectionInstance) {
  const category = instance.provider_key.split('/')[0];
  const presenter = presenters.find((candidate) => candidate.integrationType === category);
  if (presenter) return presenter.widget(instance);
  return {
    rows: [],
    remaining: 0,
    remainingLabel: 'more results',
    status: 'Waiting for information',
    checked: null,
    icon: undefined,
  };
}

/** Resolve setup presentation through the registered capability, including provider key links. */
export function setupPresentation(provider: ProviderDefinition) {
  const presenter = presenters.find((candidate) => candidate.contractVersion === provider.capability);
  if (!presenter) throw new Error('The integration capability is not installed.');
  return presenter.setup(provider);
}
