import { Locator, Page } from '@playwright/test';

/*
 * Page objects: the only file that knows how screens are built. Specs describe user flows
 * and wire contracts; when the UI is redesigned only this file changes (ADR-002).
 * Current implementation: legacy UI (before spec 007 redesign).
 */
export const ui = {
  login: {
    signInButton: (page: Page): Locator => page.getByRole('button', { name: /sign in with google/i }),
  },
  shell: {
    signOut: async (page: Page) => page.locator('.pi-sign-out').click(),
  },
  home: {
    path: '/groupdashboard',
    groupTitles: (page: Page): Locator => page.locator('.dashboard-title'),
    tile: (page: Page, group: string, label: string): Locator =>
      page.locator('.dashboard-group', { has: page.locator('.dashboard-title', { hasText: group }) }).locator('.entity-card', { has: page.locator('.entity-title', { hasText: new RegExp(`^${label}$`) }) }).first(),
    toggleControl: (tile: Locator): Locator => tile.locator('.icon-wrapper.clickable'),
  },
  automations: {
    path: '/viewer',
    row: (page: Page, name: string): Locator => page.locator('.hover-row', { hasText: name }),
    open: async (page: Page, name: string) => page.locator('.hover-row', { hasText: name }).getByText(name, { exact: true }).click(),
  },
  editor: {
    path: (id: string) => `/editor/${id}`,
    save: async (page: Page) => page.getByRole('button', { name: 'Save', exact: true }).first().click(),
  },
  devices: {
    listPath: '/devicelist',
    listRows: (page: Page): Locator => page.locator('tbody tr'),
    openFromList: async (page: Page, name: string) => page.getByRole('link', { name, exact: true }).click(),
    pageTitle: (page: Page): Locator => page.locator('.text-3xl'),
  },
};
