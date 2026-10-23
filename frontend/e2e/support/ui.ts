import { Locator, Page } from '@playwright/test';

/*
 * Page objects: the only file that knows how screens are built. Specs describe user flows
 * and wire contracts; when the UI is redesigned only this file changes (ADR-002).
 * Current implementation: spec 007 Hearth + Panel UI.
 */

/** Opens the navigation drawer on narrow screens; no-op when the sidebar is visible. */
async function openNav(page: Page) {
  // AppShell is deferred: absence of its menu is not evidence of a desktop sidebar.
  const nav = page.getByRole('navigation', { name: 'Main', includeHidden: true });
  await nav.waitFor({ state: 'attached' });
  const menu = page.getByRole('button', { name: 'Open navigation' });
  if (await menu.isVisible() && await menu.getAttribute('aria-expanded') !== 'true') await menu.click();
  await nav.waitFor({ state: 'visible' });
}

export const ui = {
  history: {
    choose: async (page: Page, name: 'Entity' | 'Period', label: string) => {
      const group = page.getByRole('region', { name: 'Device history' }).getByRole('group', { name, exact: true });
      await group.waitFor({ state: 'visible' });
      const picker = group.getByRole('combobox');
      if (await picker.isVisible()) await picker.selectOption({ label });
      else await group.getByRole('button', { name: label, exact: true }).click();
    },
    selected: async (page: Page, name: 'Entity' | 'Period') => {
      const group = page.getByRole('region', { name: 'Device history' }).getByRole('group', { name, exact: true });
      await group.waitFor({ state: 'visible' });
      const picker = group.getByRole('combobox');
      return await picker.isVisible() ? picker.locator('option:checked').textContent()
        : group.getByRole('button', { pressed: true }).textContent();
    },
  },
  login: {
    signInButton: (page: Page): Locator => page.getByRole('button', { name: /sign in with google/i }),
  },
  shell: {
    openNav,
    nav: (page: Page): Locator => page.getByRole('navigation', { name: 'Main' }),
    signOut: async (page: Page) => {
      await openNav(page);
      await page.getByRole('navigation', { name: 'Main' }).getByRole('button', { name: 'Sign out' }).click();
    },
  },
  home: {
    path: '/groupdashboard',
    groupTitles: (page: Page): Locator => page.locator('.nh-area-head h2'),
    area: (page: Page, group: string): Locator => page.locator(`section[data-area="${group}"]`),
    tile: (page: Page, group: string, label: string): Locator =>
      page.locator(`section[data-area="${group}"] .nh-tile`, { has: page.locator('.nh-tile-label', { hasText: new RegExp(label, 'i') }) }).first(),
    toggleControl: (tile: Locator): Locator => tile.locator('button.nh-tile-ic'),
    editLayout: async (page: Page) => page.getByRole('button', { name: 'Edit layout' }).click(),
    done: async (page: Page) => page.getByRole('button', { name: 'Done' }).click(),
    moveLater: (page: Page, group: string): Locator => page.getByRole('button', { name: `Move ${group} later` }),
    dragHandle: (page: Page, group: string): Locator => page.locator(`section[data-area="${group}"] .nh-area-handle`),
  },
  overview: {
    path: '/overview',
    activitySwitch: (page: Page): Locator => page.getByRole('switch', { name: 'Record recent activity' }),
    activityRows: (page: Page): Locator => page.locator('.nh-feed-row'),
  },
  panel: {
    path: '/panel',
    roomTabs: (page: Page): Locator => page.getByRole('tablist', { name: 'Rooms' }).getByRole('tab'),
    exit: (page: Page): Locator => page.getByRole('link', { name: 'Exit panel', exact: true }),
  },
  automations: {
    path: '/viewer',
    row: (page: Page, name: string): Locator => page.locator('.nh-auto', { hasText: name }),
    open: async (page: Page, name: string) => page.locator('.nh-auto-name', { hasText: name }).click(),
  },
  editor: {
    path: (id: string) => `/editor/${id}`,
    save: async (page: Page) => page.getByRole('button', { name: 'Save', exact: true }).click(),
    enabled: (page: Page) => page.getByRole('switch', { name: 'Automation on' }),
    addTrigger: (page: Page) => page.getByRole('button', { name: 'Add trigger', exact: true }),
    removeTrigger: async (page: Page) => {
      await page.getByRole('button', { name: 'Delete trigger', exact: true }).click();
      await page.getByRole('dialog', { name: 'Delete trigger 1?' }).getByRole('button', { name: 'Delete trigger', exact: true }).click();
    },
    when: (page: Page) => page.getByRole('combobox', { name: 'Property that starts this trigger', exact: true }),
    addCondition: (page: Page) => page.getByRole('button', { name: 'Value condition', exact: true }),
    conditionProperty: (page: Page) => page.getByRole('combobox', { name: 'Property', exact: true }),
    compareWith: (page: Page) => page.getByRole('combobox', { name: 'Compare with', exact: true }),
    addAction: (page: Page) => page.getByRole('button', { name: 'Add action', exact: true }),
    actionDevice: (page: Page) => page.getByRole('combobox', { name: 'Action 1 device', exact: true }),
    addValue: (page: Page) => page.getByRole('button', { name: 'Add value', exact: true }),
    propertyToSet: (page: Page) => page.getByRole('combobox', { name: 'Property to set', exact: true }),
    newValue: (page: Page) => page.getByRole('combobox', { name: 'New value', exact: true }),
    numericValue: (page: Page) => page.getByRole('spinbutton', { name: 'New value', exact: true }),
    summary: (page: Page) => page.locator('.nh-sentence'),
  },
  devices: {
    path: (id: string) => `/devicepage/${id}`,
    state: (page: Page) => page.getByRole('switch', { name: 'State', exact: true }),
    listPath: '/devicelist',
    listRows: (page: Page): Locator => page.locator('tbody tr'),
    openFromList: async (page: Page, name: string) => page.getByRole('link', { name: new RegExp(`^${name}`) }).first().click(),
    pageTitle: (page: Page): Locator => page.locator('h1'),
  },
};
