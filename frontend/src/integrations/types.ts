import type { Component } from 'vue';
import type { ConnectionInstance, IntegrationResource, ProviderDefinition } from '@/types/integrations.type';
export type { IntegrationResource } from '@/types/integrations.type';
import type { WasteReference } from './waste/reference';
import type { WeatherReference } from './weather/reference';

/** Installed source contracts are composed here; Home treats their identity as opaque. */
export type IntegrationSourceReference = WasteReference | WeatherReference;

/** A capability-owned icon component and its presentation-only inputs. */
export interface SourceIcon {
  component: Component;
  props: Record<string, unknown>;
}
/** One source summary row; values and icons belong to the capability presenter. */
export interface SourceRow {
  id: string;
  title: string;
  value: string;
  colour: string | null;
  icon?: SourceIcon;
}
/** Shared Home/Panel view model; it owns neither transport nor saved placement. */
export interface IntegrationSourcePresentation {
  pickerLabel?: string;
  selectable?: boolean;
  detail?: string;
  icon?: SourceIcon;
  reference: IntegrationSourceReference;
  title: string;
  value: string;
  origin: string;
  status: string;
  checked: string;
  colour: string | null;
  warning: boolean;
  error?: boolean;
  rows: SourceRow[];
}
/** Domain presentation boundary consumed by generic Home and Connections components. */
export interface IntegrationSourceKind {
  title: string;
  requiresService: boolean;
}
export interface IntegrationPresenter {
  setup(provider: ProviderDefinition): IntegrationSetupPresentation;
  contractVersion: string;
  integrationType: string;
  widget(instance: ConnectionInstance): IntegrationWidget;
  sourceKinds: Record<string, IntegrationSourceKind>;
  sources(instance: ConnectionInstance, resource: IntegrationResource): IntegrationSourcePresentation[];
  icon(kind: string, colour: string | null): SourceIcon;
  results: Component;
}

/** Compact connection presentation supplied by a capability, with opaque icons. */
export interface IntegrationWidgetRow {
  id: string;
  name: string;
  colour: string | null;
  nextDateISO: string | null;
  date: string;
  icon?: SourceIcon;
}
export interface IntegrationWidget {
  rows: IntegrationWidgetRow[];
  remaining: number;
  remainingLabel: string;
  status: string;
  checked: string | null;
  icon?: SourceIcon;
}

/** Capability-owned setup copy and input hints consumed by the generic wizard. */
export interface IntegrationSetupPresentation {
  title: string;
  sectionTitle: string;
  defaultName: string;
  searchLabel: string;
  searchButton: string;
  choiceLabel: string;
  choicePlaceholder: string;
  searchMaxLength: number;
  autocomplete: string;
  notice: string;
  keyURL?: string;
}
