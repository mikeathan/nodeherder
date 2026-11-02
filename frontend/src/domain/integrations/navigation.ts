import type { LocationQuery, RouteLocationRaw } from 'vue-router';
import { RouteName } from '@/types/router';

export type IntegrationSection = 'results' | 'setup' | 'operations' | 'help';

/** Allowlisted internal return context; imported/provider query strings cannot create redirects. */
export function integrationReturnTarget(query: LocationQuery): { to: RouteLocationRaw; label: string } {
  if (query.returnTo === '/panel') return { to: { name: RouteName.Panel }, label: 'Panel' };
  if (query.returnTo === '/groupdashboard') return { to: { name: RouteName.GroupDashboard }, label: 'Home' };
  return { to: { name: RouteName.Integrations }, label: 'Integrations' };
}

/** Explicit navigation context; a shared provider page never infers an instance from a list. */
export interface IntegrationNavigationContext {
  providerKey: string;
  instanceId?: string;
  query?: LocationQuery;
}

export interface IntegrationSectionLink {
  section: IntegrationSection;
  label: string;
  to: RouteLocationRaw;
}

const integrationPageNames: readonly unknown[] = [
  RouteName.Integration,
  RouteName.IntegrationSetup,
  RouteName.IntegrationProvider,
  RouteName.IntegrationHelp,
  RouteName.IntegrationNotFound,
];

/** Read only a single bounded instance reference; repeated/missing query values have no context. */
export function integrationInstanceId(query: LocationQuery): string | undefined {
  const value = query.instance;
  if (typeof value !== 'string' || !value || value.length > 256) return undefined;
  return value;
}

/** Select the saved-instance view from its URL; unsupported hashes safely show Results. */
export function integrationDetailSection(hash: string): 'results' | 'setup' {
  return hash === '#setup' ? 'setup' : 'results';
}

/** Remount instance/provider page owners when their path identity changes, including legacy
 * aliases. Hash/query-only section changes retain the current owner and its in-progress inputs.
 */
export function integrationViewKey(name: unknown, path: string): string | undefined {
  return integrationPageNames.includes(name) ? path : undefined;
}

/** Build authenticated named links without performing requests or changing shared state.
 * Provider/Help links retain editor query context. Detail links retain only allowlisted return
 * context and carry an explicit instance
 * and section hash, so reload/history work without restoring a local tab variable.
 */
export function integrationSectionLinks(context: IntegrationNavigationContext): IntegrationSectionLink[] {
  const links: IntegrationSectionLink[] = [];
  if (context.instanceId) {
    const params = { id: context.instanceId };
    const returnTo = context.query?.returnTo;
    const returning = returnTo === '/panel' || returnTo === '/groupdashboard' ? { query: { returnTo } } : {};
    links.push(
      {
        section: 'results',
        label: 'Results',
        to: { name: RouteName.Integration, params, hash: '#results', ...returning },
      },
      { section: 'setup', label: 'Setup', to: { name: RouteName.Integration, params, hash: '#setup', ...returning } }
    );
  }
  const parts = context.providerKey.split('/');
  if (parts.length !== 2 || !parts.every(Boolean)) return links;
  const [type, provider] = parts;
  const params = { type, provider };
  const query = { ...context.query };
  // Never mutate the router's query object or propagate an ambiguous instance reference.
  delete query.instance;
  if (context.instanceId) query.instance = context.instanceId;
  links.push(
    {
      section: 'operations',
      label: 'Configure operations',
      to: { name: RouteName.IntegrationProvider, params, query },
    },
    { section: 'help', label: 'Help', to: { name: RouteName.IntegrationHelp, params, query } }
  );
  return links;
}
