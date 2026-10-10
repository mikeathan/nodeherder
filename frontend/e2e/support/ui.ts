import { Locator, Page } from '@playwright/test';

/*
 * Page objects: the only file that knows how screens are built. Specs describe user flows
 * and wire contracts; when the UI is redesigned only this file changes (ADR-002).
 * Current implementation: spec 007 Hearth + Panel UI.
 */

/** Opens the navigation drawer on narrow screens; no-op when the sidebar is visible. */
async function openNav(page: Page) {
  const menu = page.getByRole('button', { name: 'Open navigation' });
  if (await menu.isVisible()) await menu.click();
}

export const ui = {
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
  },
  devices: {
    listPath: '/devicelist',
    listRows: (page: Page): Locator => page.locator('tbody tr'),
    openFromList: async (page: Page, name: string) => page.getByRole('link', { name: new RegExp(`^${name}`) }).first().click(),
    pageTitle: (page: Page): Locator => page.locator('h1'),
  },
};
