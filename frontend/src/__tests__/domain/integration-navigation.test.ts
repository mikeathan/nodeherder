import {
  integrationDetailSection,
  integrationInstanceId,
  integrationReturnTarget,
  integrationSectionLinks,
  integrationViewKey,
} from '@/domain/integration-navigation';
import { RouteName } from '@/types/router';

test('instance sections retain explicit identity while provider links preserve editor context', () => {
  const query = { instance: 'instance-a', new: '1', from: 'waste_collection/original' };
  const links = integrationSectionLinks({ providerKey: 'waste_collection/copied', instanceId: 'instance-a', query });
  expect(links.map((link) => link.section)).toEqual(['results', 'setup', 'operations', 'help']);
  expect(links[1].to).toEqual({ name: RouteName.Integration, params: { id: 'instance-a' }, hash: '#setup' });
  expect(links[3].to).toEqual({
    name: RouteName.IntegrationHelp,
    params: { type: 'waste_collection', provider: 'copied' },
    query,
  });
  expect(query).toEqual({ instance: 'instance-a', new: '1', from: 'waste_collection/original' });
});

test('direct shared provider entry has no guessed instance or ambiguous query reference', () => {
  const links = integrationSectionLinks({ providerKey: 'waste_collection/provider', query: { instance: ['a', 'b'] } });
  expect(links.map((link) => link.section)).toEqual(['operations', 'help']);
  expect(links[0].to).toMatchObject({ query: {} });
  expect(integrationInstanceId({ instance: ['a', 'b'] })).toBeUndefined();
  expect(integrationInstanceId({ instance: null })).toBeUndefined();
  expect(integrationInstanceId({ instance: 'a' })).toBe('a');
  expect(integrationInstanceId({ instance: 'a'.repeat(257) })).toBeUndefined();
});

test('detail hashes survive reload and unknown sections fall back to Results', () => {
  expect(integrationDetailSection('#setup')).toBe('setup');
  expect(integrationDetailSection('#results')).toBe('results');
  expect(integrationDetailSection('#invalid')).toBe('results');
  expect(integrationDetailSection('')).toBe('results');
});

test('canonical and legacy integration paths fence page owners without affecting other views', () => {
  expect(integrationViewKey(RouteName.IntegrationProvider, '/integrations/providers/type/one')).toBe(
    '/integrations/providers/type/one'
  );
  expect(integrationViewKey(RouteName.Integration, '/connections/second')).toBe('/connections/second');
  expect(integrationViewKey(RouteName.GroupDashboard, '/groupdashboard')).toBeUndefined();
});

test('Home and panel return links survive section navigation without accepting redirects', () => {
  for (const returnTo of ['/panel', '/groupdashboard']) {
    const links = integrationSectionLinks({
      providerKey: 'waste/provider',
      instanceId: 'one',
      query: { returnTo, secret: 'ignored' },
    });
    expect(links[0].to).toMatchObject({ query: { returnTo } });
    expect(links[1].to).toMatchObject({ query: { returnTo } });
    expect(integrationReturnTarget({ returnTo }).label).toBe(returnTo === '/panel' ? 'Panel' : 'Home');
  }
  for (const returnTo of ['https://elsewhere.test', '//elsewhere.test', ['/panel', '/groupdashboard'], null]) {
    expect(integrationReturnTarget({ returnTo }).to).toEqual({ name: RouteName.Integrations });
    expect(
      integrationSectionLinks({ providerKey: 'waste/provider', instanceId: 'one', query: { returnTo } })[0].to
    ).not.toHaveProperty('query');
  }
});
