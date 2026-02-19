# Frontend Redesign Implementation Plan

# Goal

Refactor the frontend to eliminate styling debt (remove Bootstrap), implement a robust CSS-only responsive layout, and add a global theme toggle. This aligns with **Strategy 1: Cleanup & Standardize**.

## User Review Required

> [!IMPORTANT]◊
> This refactor will exclusively use **PrimeFlex** and **CSS** for layout. Bootstrap will be completely removed. Any specific Bootstrap utility classes used deep in components might need manual adjustment, though a broad pass will be made.

## Proposed Changes

### Phase 1: Standardization & Cleanup

#### [MODIFY] [package.json](file:///Users/mikeathan/dev/nodeherder/frontend/package.json)

- Remove `bootstrap` dependency.
- Ensure `primeflex` and `primevue` are up to date.

#### [MODIFY] [src/main.ts](file:///Users/mikeathan/dev/nodeherder/frontend/src/main.ts)

- Remove `import 'bootstrap/dist/css/bootstrap.min.css';`
- Remove `import 'bootstrap';`

### Phase 2: Layout Engine

#### [MODIFY] [src/components/layout/MainLayout.vue](file:///Users/mikeathan/dev/nodeherder/frontend/src/components/layout/MainLayout.vue)

- **Remove**: `drawerWidth` state, `handleDrawerWidthChanged`, `style="{ marginLeft: ... }"` binding.
- **Implement**:
  - A standard administrative shell layout using CSS Flexbox.
  - **Sidebar**: Fixed width on desktop (`w-64`), hidden on mobile.
  - **Header**: Sticky top.
  - **Content**: Scrollable area filling remaining space.
- **Mobile**: Use `PrimeVue`'s `Sidebar` (Drawer) component for the navigation menu on small screens.

#### [MODIFY] [src/App.vue](file:///Users/mikeathan/dev/nodeherder/frontend/src/App.vue)

- Ensure root level container is suitable for full-screen layout (e.g., `h-screen`, `overflow-hidden`).

### Phase 3: Theming

#### [NEW] [src/services/theme.service.ts](file:///Users/mikeathan/dev/nodeherder/frontend/src/services/theme.service.ts)

- Simple state management for the current theme (light/dark).
- Logic to toggle PrimeVue themes (if using PrimeVue theming API) or toggle a root class for CSS variables.

#### [MODIFY] [src/components/controls/NavigationBar.vue](file:///Users/mikeathan/dev/nodeherder/frontend/src/components/controls/NavigationBar.vue)

- Add a **Theme Toggle Button** (Sun/Moon icon).

#### [MODIFY] [src/assets/styles/variables.css](file:///Users/mikeathan/dev/nodeherder/frontend/src/assets/styles/variables.css) (New File or verify existing)

- Define semantic CSS variables:
  - `--surface-ground`
  - `--surface-card`
  - `--text-color`
  - `--primary-color`
- Ensure these variables map correctly to Light and Dark values.

## Verification Plan

### Manual Verification

1.  **Dependency Check**: Ensure app builds without Bootstrap errors.
2.  **Layout Stress Test**:
    - **Desktop**: Resize window. Sidebar should remain fixed, content should resize.
    - **Mobile**: Sidebar should disappear. Hamburger menu should open a drawer.
3.  **Theme Test**:
    - Click toggle. Backgrounds and text colors should invert instantly.
    - Reload page. Theme preference should persist (optional, but good for UX).
