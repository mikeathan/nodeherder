/*
 * PrimeVue preset mapped onto the --nh-* design tokens (ADR-001). PrimeVue keeps behaviour and
 * accessibility; every colour it paints resolves to the active theme's CSS variables, so a
 * theme, accent or mode change restyles PrimeVue widgets with no JavaScript.
 */
import { definePreset } from '@primevue/themes';
import base, { type Preset } from '@primevue/themes/aura/base';
import badge from '@primevue/themes/aura/badge';
import button from '@primevue/themes/aura/button';
import card from '@primevue/themes/aura/card';
import dialog from '@primevue/themes/aura/dialog';
import drawer from '@primevue/themes/aura/drawer';
import floatlabel from '@primevue/themes/aura/floatlabel';
import iconfield from '@primevue/themes/aura/iconfield';
import inputnumber from '@primevue/themes/aura/inputnumber';
import inputtext from '@primevue/themes/aura/inputtext';
import menu from '@primevue/themes/aura/menu';
import message from '@primevue/themes/aura/message';
import ripple from '@primevue/themes/aura/ripple';
import select from '@primevue/themes/aura/select';
import slider from '@primevue/themes/aura/slider';
import tag from '@primevue/themes/aura/tag';
import textarea from '@primevue/themes/aura/textarea';
import toast from '@primevue/themes/aura/toast';
import toggleswitch from '@primevue/themes/aura/toggleswitch';
import tooltip from '@primevue/themes/aura/tooltip';
import virtualscroller from '@primevue/themes/aura/virtualscroller';

const mix = (pct: number, other: string) => `color-mix(in srgb, var(--nh-accent) ${pct}%, ${other})`;

const primaryPalette = {
  50: mix(10, 'white'),
  100: mix(20, 'white'),
  200: mix(40, 'white'),
  300: mix(60, 'white'),
  400: mix(80, 'white'),
  500: 'var(--nh-accent)',
  600: mix(85, 'black'),
  700: mix(70, 'black'),
  800: mix(55, 'black'),
  900: mix(40, 'black'),
  950: mix(25, 'black'),
};

/** Shared semantic colours: identical for both modes because the variables already switch. */
const semantic = {
  primary: {
    color: 'var(--nh-accent)',
    contrastColor: 'var(--nh-accent-contrast)',
    hoverColor: mix(85, 'var(--nh-text)'),
    activeColor: mix(75, 'var(--nh-text)'),
  },
  highlight: {
    background: 'var(--nh-accent-soft)',
    focusBackground: 'var(--nh-accent-soft)',
    color: 'var(--nh-text)',
    focusColor: 'var(--nh-text)',
  },
  mask: { background: 'rgb(0 0 0 / 0.45)', color: 'var(--nh-text)' },
  formField: {
    background: 'var(--nh-surface)',
    disabledBackground: 'var(--nh-surface-2)',
    filledBackground: 'var(--nh-surface-2)',
    filledHoverBackground: 'var(--nh-surface-2)',
    filledFocusBackground: 'var(--nh-surface)',
    borderColor: 'var(--nh-border-strong)',
    hoverBorderColor: 'var(--nh-text-muted)',
    focusBorderColor: 'var(--nh-accent)',
    invalidBorderColor: 'var(--nh-danger)',
    color: 'var(--nh-text)',
    disabledColor: 'var(--nh-text-muted)',
    placeholderColor: 'var(--nh-text-muted)',
    invalidPlaceholderColor: 'var(--nh-danger)',
    floatLabelColor: 'var(--nh-text-muted)',
    floatLabelFocusColor: 'var(--nh-accent)',
    floatLabelActiveColor: 'var(--nh-text-muted)',
    floatLabelInvalidColor: 'var(--nh-danger)',
    iconColor: 'var(--nh-text-muted)',
    shadow: 'none',
  },
  text: {
    color: 'var(--nh-text)',
    hoverColor: 'var(--nh-text)',
    mutedColor: 'var(--nh-text-muted)',
    hoverMutedColor: 'var(--nh-text)',
  },
  content: {
    background: 'var(--nh-surface)',
    hoverBackground: 'var(--nh-surface-2)',
    borderColor: 'var(--nh-border)',
    color: 'var(--nh-text)',
    hoverColor: 'var(--nh-text)',
  },
  overlay: {
    select: { background: 'var(--nh-surface)', borderColor: 'var(--nh-border-strong)', color: 'var(--nh-text)' },
    popover: { background: 'var(--nh-surface)', borderColor: 'var(--nh-border-strong)', color: 'var(--nh-text)' },
    modal: { background: 'var(--nh-surface)', borderColor: 'var(--nh-border-strong)', color: 'var(--nh-text)' },
  },
  list: {
    option: {
      focusBackground: 'var(--nh-surface-2)',
      selectedBackground: 'var(--nh-accent-soft)',
      selectedFocusBackground: 'var(--nh-accent-soft)',
      color: 'var(--nh-text)',
      focusColor: 'var(--nh-text)',
      selectedColor: 'var(--nh-text)',
      selectedFocusColor: 'var(--nh-text)',
      icon: { color: 'var(--nh-text-muted)', focusColor: 'var(--nh-text)' },
    },
    optionGroup: { background: 'transparent', color: 'var(--nh-text-muted)' },
  },
  navigation: {
    item: {
      focusBackground: 'var(--nh-surface-2)',
      activeBackground: 'var(--nh-accent-soft)',
      color: 'var(--nh-text)',
      focusColor: 'var(--nh-text)',
      activeColor: 'var(--nh-accent)',
      icon: { color: 'var(--nh-text-muted)', focusColor: 'var(--nh-text)', activeColor: 'var(--nh-accent)' },
    },
    submenuLabel: { background: 'transparent', color: 'var(--nh-text-muted)' },
    submenuIcon: { color: 'var(--nh-text-muted)', focusColor: 'var(--nh-text)', activeColor: 'var(--nh-accent)' },
  },
};

/**
 * Aura tokens for the PrimeVue components this app renders. The stock
 * `@primevue/themes/aura` entry imports every component's tokens, including unused ones,
 * so importing only the used subset keeps the initial JavaScript smaller. Add the matching
 * token module when adding a new PrimeVue component.
 */
const AuraSubset = {
  ...base,
  components: {
    badge,
    button,
    card,
    dialog,
    drawer,
    floatlabel,
    iconfield,
    inputnumber,
    inputtext,
    menu,
    message,
    ripple,
    select,
    slider,
    tag,
    textarea,
    toast,
    toggleswitch,
    tooltip,
    virtualscroller,
  },
  // `base` declares primitive/semantic as `unknown`, which Preset's index signature rejects;
  // the runtime shape is the same base token object the stock Aura preset spreads.
} as unknown as Preset;

export const NodeHerderPreset = definePreset(AuraSubset, {
  semantic: {
    primary: primaryPalette,
    borderRadius: { none: '0', xs: 'var(--nh-radius-s)', sm: 'var(--nh-radius-s)', md: 'var(--nh-radius-m)', lg: 'var(--nh-radius-l)', xl: 'var(--nh-radius-l)' },
    focusRing: { width: '2px', style: 'solid', color: 'var(--nh-focus)', offset: '2px', shadow: 'none' },
    colorScheme: {
      light: {
        surface: {
          0: 'var(--nh-surface)', 50: 'var(--nh-bg)', 100: 'var(--nh-surface-2)', 200: 'var(--nh-border)',
          300: 'var(--nh-border-strong)', 400: 'var(--nh-text-muted)', 500: 'var(--nh-text-muted)', 600: 'var(--nh-text-muted)',
          700: 'var(--nh-text)', 800: 'var(--nh-text)', 900: 'var(--nh-text)', 950: 'var(--nh-text)',
        },
        ...semantic,
      },
      dark: {
        // Aura's dark scheme reads high surface steps as backgrounds and low steps as text.
        surface: {
          0: 'var(--nh-text)', 50: 'var(--nh-text)', 100: 'var(--nh-text)', 200: 'var(--nh-text-muted)',
          300: 'var(--nh-text-muted)', 400: 'var(--nh-text-muted)', 500: 'var(--nh-text-muted)', 600: 'var(--nh-border-strong)',
          700: 'var(--nh-border)', 800: 'var(--nh-surface-2)', 900: 'var(--nh-surface)', 950: 'var(--nh-bg)',
        },
        ...semantic,
      },
    },
  },
});

export const PRIMEVUE_THEME_OPTIONS = {
  darkModeSelector: '[data-mode="dark"]',
  cssLayer: false,
} as const;
