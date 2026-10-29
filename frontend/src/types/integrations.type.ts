import type { WeatherForecast, WeatherMapping, WeatherContractVersion } from './weather.type';
import type { WASTE_SCHEDULE_TIMEZONE } from '@/config/integrations';
export interface ConnectionFieldError {
  path: string;
  message: string;
  line?: number;
  column?: number;
}
export type ConnectionFailure = {
  code: string;
  message: string;
  retry_at?: string;
  field_errors?: ConnectionFieldError[];
};
export interface ConnectionValueReference {
  ref: string;
}
export type ConnectionValue = string | number | boolean | null | ConnectionValueReference;
export type IntegrationSchemaVersion = 'nodeherder.integration/v1';
export type IntegrationType = 'waste_collection' | 'weather';
export type WasteContractVersion = 'waste_schedule/v1';
export type IntegrationCapability = WasteContractVersion | WeatherContractVersion;
export type ProviderRequestMethod = 'GET' | 'POST';
export type ProviderSetupType = 'search_select' | 'location_select';
export type ProviderEditorMode = 'form' | 'yaml';
export type ProviderTestMode = 'offline' | 'live';
export type IntegrationResource = WasteSchedule | WeatherForecast;

export interface ProviderCapture {
  pointer: string;
  type: string;
  when?: string;
  strip_prefix?: string;
  optional?: boolean;
}
export interface ProviderRequest {
  combine?: string;
  method: ProviderRequestMethod;
  path: ConnectionValue[];
  query?: Record<string, ConnectionValue>;
  body?: Record<string, ConnectionValue>;
  idempotency?: boolean;
  conditional?: boolean;
  response_root?: string;
  captures?: Record<string, ProviderCapture>;
}
export interface ProviderAuth {
  mode: string;
  secret_slot?: string;
}
export interface ProviderPolicy {
  poll_min_seconds: number;
  renewal_min_seconds: number;
  idempotency_window_seconds: number;
}
export interface ProviderErrorRule {
  status: number;
  code: string;
  action: string;
}
export interface ProviderErrors {
  code: string;
  rules: ProviderErrorRule[];
}
export interface ProviderInput {
  type: string;
  label: string;
  required: boolean;
}
export interface ProviderChoiceMapping {
  items: string;
  value: string;
  label: string;
  latitude?: string;
  longitude?: string;
  context?: string[];
}
export interface ProviderPostcodeChoiceMapping {
  items: string;
  label: string;
  latitude: string;
  longitude: string;
  context?: string[];
}
export interface ProviderPostcodeSearch {
  base_url: string;
  request: string;
  options: ProviderPostcodeChoiceMapping;
}
export interface ProviderSearchExpectation {
  pointer: string;
  value: string;
}
export interface ProviderSearchStep {
  base_url?: string;
  postcode?: ProviderPostcodeSearch;
  input: string;
  request: string;
  expect?: ProviderSearchExpectation;
  options: ProviderChoiceMapping;
}
export interface ProviderSelectionStep {
  input: string;
  request: string;
}
export interface ProviderStatusStep {
  request: string;
  pointer: string;
  pending: string[];
  done: string[];
  failed: string[];
}
export interface ProviderSetup {
  type: ProviderSetupType;
  search: ProviderSearchStep;
  select: ProviderSelectionStep;
  status: ProviderStatusStep;
}
export interface ProviderRefresh {
  request: string;
  renewal_request?: string;
}
export interface WasteCollectionMapping {
  items: string;
  fields: Record<string, string>;
}
export interface WasteMapping {
  metadata: Record<string, string>;
  collections: WasteCollectionMapping;
}
export type ProviderMapping = WeatherMapping | WasteMapping;

export interface ProviderDefinition {
  schema_version: IntegrationSchemaVersion;
  integration_type: IntegrationType;
  id: string;
  name: string;
  capability: IntegrationCapability;
  base_url: string;
  auth: ProviderAuth;
  policy: ProviderPolicy;
  errors: ProviderErrors;
  inputs: Record<string, ProviderInput>;
  requests: Record<string, ProviderRequest>;
  setup: ProviderSetup;
  refresh: ProviderRefresh;
  mapping: ProviderMapping;
}
export type WasteSubscription = 'required' | 'not_required' | 'unknown';
export type WasteDateState = 'known' | 'confirmed_empty' | 'unavailable';
export type WasteFreshness = 'fresh' | 'stale' | 'unknown';
export interface WasteCoverage {
  state: string;
  from: string | null;
  through: string | null;
}
export interface WasteQuality {
  address_match: string;
  evidence_granularity: string;
  date_confidence: string;
  provisional: boolean;
  source_confidence_label: string | null;
}
export interface WasteProvenance {
  provider_id: string;
  provider_version: string;
  profile_revision: number;
  binding_revision: number;
  council_gss: string | null;
  source_url: string | null;
}
export interface WasteService {
  service_id: string;
  source_service_key: string | null;
  name: string;
  waste_type: string;
  source_waste_type: string | null;
  bin_colour: string | null;
  container: string | null;
  subscription: WasteSubscription;
  date_state: WasteDateState;
  dates: string[];
  coverage: WasteCoverage;
}
export interface WasteSchedule {
  contract_version: WasteContractVersion;
  instance_id: string;
  resource_id: string;
  revision: number;
  kind: string;
  timezone: typeof WASTE_SCHEDULE_TIMEZONE;
  captured_at: string;
  last_checked_at: string;
  source_retrieved_at: string | null;
  fresh_until: string | null;
  evaluated_at: string;
  freshness: WasteFreshness;
  quality: WasteQuality;
  provenance: WasteProvenance;
  representative_consent: string;
  notices: string[];
  services: WasteService[];
}
export interface ConnectionCatalogType {
  id: string;
  name: string;
  description: string;
}
export interface ConnectionCatalogProvider {
  key: string;
  id: string;
  name: string;
  integration_type: string;
  origin: string;
  available: boolean;
  disk_error?: string;
}
export interface ConnectionCatalog {
  available: boolean;
  availability_error: ConnectionFailure | null;
  types: ConnectionCatalogType[];
  providers: ConnectionCatalogProvider[];
}
export interface ConnectionChoice {
  id: string;
  label: string;
}
export interface ConnectionRun {
  operation_id: string;
  kind: string;
  outcome: string;
  message: string;
  at: string;
}
export interface ConnectionSetup {
  preview?: IntegrationResource;
  id: string;
  revision: number;
  etag: string;
  provider_key: string;
  status: string;
  choices: ConnectionChoice[];
  postcode?: string;
  instance_id?: string;
  operation_id?: string;
  expires_at: string;
  credential_configured: boolean;
  last_error: ConnectionFailure | null;
}
export interface ConnectionInstance {
  credential_slots: string[];
  runs: ConnectionRun[];
  id: string;
  name: string;
  provider_key: string;
  revision: number;
  etag: string;
  enabled: boolean;
  status: string;
  postcode: string;
  operation_id?: string;
  credential_configured: boolean;
  resources: IntegrationResource[];
  last_error: ConnectionFailure | null;
  updated_at: string;
}
/** One shared placement per configured connection; array position defines display order. */
export interface ConnectionBoardEntry {
  instance_id: string;
  hidden: boolean;
}
export interface ConnectionBoard {
  etag: string;
  entries: ConnectionBoardEntry[];
}
export interface ConnectionOperation {
  id: string;
  status: string;
  kind: string;
  instance_id?: string;
  setup_id?: string;
  retry_at: string;
  expires_at: string;
  last_error: ConnectionFailure | null;
}
export interface ConnectionDefinition {
  shipped_yaml?: string;
  key: string;
  yaml: string;
  document: ProviderDefinition;
  content_hash: string;
  active_hash: string;
  revision: number;
  etag: string;
  status: string;
  disk_error?: string;
  update_available: boolean;
}
export interface CredentialReview {
  credential_slot: string;
  instance_id: string;
  revision: number;
  credential_revision: number;
}
export interface DestinationReview {
  origin: string;
  reassociate: CredentialReview[];
}
export interface DefinitionValidation {
  yaml: string;
  document: ProviderDefinition;
  content_hash: string;
  origin: string;
  previous_origin: string;
  affected_instances: CredentialReview[];
}
export interface ProviderDraft {
  yaml: string;
  etag: string;
  mode: ProviderEditorMode;
  dirty: boolean;
  scroll: number;
  focus?: string;
  document?: ProviderDefinition;
}
