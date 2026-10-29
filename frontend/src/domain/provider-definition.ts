import type { ConnectionValue, ProviderDefinition, ProviderRequest } from '@/types/integrations.type';

const identifier = /^[a-z][a-z0-9_]{0,63}$/;

/** Rename a request and every recipe link atomically; the input document stays untouched. */
export function renameProviderRequest(
  definition: ProviderDefinition,
  oldName: string,
  name: string
): ProviderDefinition {
  if (!identifier.test(name) || name in definition.requests || !(oldName in definition.requests)) return definition;
  const link = <T extends { request: string }>(step: T): T => ({
    ...step,
    request: step.request === oldName ? name : step.request,
  });
  return {
    ...definition,
    requests: Object.fromEntries(
      Object.entries(definition.requests).map(([key, value]) => [key === oldName ? name : key, { ...value, combine: value.combine === oldName ? name : value.combine }])
    ),
    setup: {
      ...definition.setup,
      search: {
        ...link(definition.setup.search),
        ...(definition.setup.search.postcode && { postcode: link(definition.setup.search.postcode) }),
      },
      select: link(definition.setup.select),
      status: link(definition.setup.status),
    },
    refresh: {
      ...link(definition.refresh),
      renewal_request: definition.refresh.renewal_request === oldName ? name : definition.refresh.renewal_request,
    },
  };
}

/** Saved state names are global: rename all producers and typed references together. */
export function renameProviderCapture(
  definition: ProviderDefinition,
  oldName: string,
  name: string
): ProviderDefinition {
  const requests = Object.values(definition.requests);
  if (!requests.some((request) => oldName in (request.captures ?? {}))) return definition;
  if (!identifier.test(name) || requests.some((request) => name in (request.captures ?? {}))) return definition;
  const replace = (value: ConnectionValue): ConnectionValue =>
    value !== null && typeof value === 'object' && value.ref === `state.${oldName}` ? { ref: `state.${name}` } : value;
  const fields = (values: Record<string, ConnectionValue> | undefined) =>
    values && Object.fromEntries(Object.entries(values).map(([key, value]) => [key, replace(value)]));
  return {
    ...definition,
    requests: Object.fromEntries(
      Object.entries(definition.requests).map(([id, request]) => [
        id,
        {
          ...request,
          path: request.path.map(replace),
          query: fields(request.query),
          body: fields(request.body),
          captures:
            request.captures &&
            Object.fromEntries(
              Object.entries(request.captures).map(([key, value]) => [key === oldName ? name : key, value])
            ),
        },
      ])
    ),
  };
}

/** The next unused key keeps repeated Add actions deterministic and collision free. */
export function nextDefinitionKey(values: Record<string, unknown>, prefix: string): string {
  let suffix = 1;
  while (`${prefix}_${suffix}` in values) suffix++;
  return `${prefix}_${suffix}`;
}

export function requestPathFields(request: ProviderRequest): Record<string, ConnectionValue> {
  return Object.fromEntries(request.path.map((value, index) => [`segment_${index + 1}`, value]));
}
