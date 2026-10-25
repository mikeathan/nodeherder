export type ConnectionFailure = {
  code: string;
  message: string;
  retry_at?: string;
  field_errors?: { path: string; message: string; line?: number; column?: number }[];
};
export type ConnectionValue = string | number | boolean | null | { ref: string };
export interface ProviderRequest {
  method: 'GET' | 'POST';
  path: ConnectionValue[];
  query?: Record<string, ConnectionValue>;
  body?: Record<string, ConnectionValue>;
  idempotency?: boolean;
  conditional?: boolean;
  response_root?: string;
  captures?: Record<
    string,
    { pointer: string; type: string; when?: string; strip_prefix?: string; optional?: boolean }
  >;
}
export interface ProviderDefinition {
  schema_version: 'nodeherder.integration/v1';
  integration_type: 'waste_collection';
  id: string;
  name: string;
  capability: 'waste_schedule/v1';
  base_url: string;
  auth: { mode: string; secret_slot?: string };
  policy: { poll_min_seconds: number; renewal_min_seconds: number; idempotency_window_seconds: number };
  errors: { code: string; rules: { status: number; code: string; action: string }[] };
  inputs: Record<string, { type: string; label: string; required: boolean }>;
  requests: Record<string, ProviderRequest>;
  setup: {
    type: 'search_select';
    search: {
      input: string;
      request: string;
      expect?: { pointer: string; value: string };
      options: { items: string; value: string; label: string };
    };
    select: { input: string; request: string };
    status: { request: string; pointer: string; pending: string[]; done: string[]; failed: string[] };
  };
  refresh: { request: string; renewal_request?: string };
  mapping: { metadata: Record<string, string>; collections: { items: string; fields: Record<string, string> } };
}
export interface WasteService {
  service_id: string;
  source_service_key: string | null;
  name: string;
  waste_type: string;
  source_waste_type: string | null;
  bin_colour: string | null;
  container: string | null;
  subscription: 'required' | 'not_required' | 'unknown';
  date_state: 'known' | 'confirmed_empty' | 'unavailable';
  dates: string[];
  coverage: { state: string; from: string | null; through: string | null };
}
export interface WasteSchedule {
  contract_version: 'waste_schedule/v1';
  instance_id: string;
  resource_id: string;
  revision: number;
  kind: string;
  timezone: 'Europe/London';
  captured_at: string;
  last_checked_at: string;
  source_retrieved_at: string | null;
  fresh_until: string | null;
  evaluated_at: string;
  freshness: 'fresh' | 'stale' | 'unknown';
  quality: {
    address_match: string;
    evidence_granularity: string;
    date_confidence: string;
    provisional: boolean;
    source_confidence_label: string | null;
  };
  provenance: {
    provider_id: string;
    provider_version: string;
    profile_revision: number;
    binding_revision: number;
    council_gss: string | null;
    source_url: string | null;
  };
  representative_consent: string;
  notices: string[];
  services: WasteService[];
}
export interface ConnectionCatalog {
  available: boolean;
  availability_error: ConnectionFailure | null;
  types: { id: string; name: string; description: string }[];
  providers: {
    key: string;
    id: string;
    name: string;
    integration_type: string;
    origin: string;
    available: boolean;
    disk_error?: string;
  }[];
}
export interface ConnectionSetup {
  preview?: WasteSchedule;
  id: string;
  revision: number;
  etag: string;
  provider_key: string;
  status: string;
  choices: { id: string; label: string }[];
  postcode?: string;
  instance_id?: string;
  operation_id?: string;
  expires_at: string;
  credential_configured: boolean;
  last_error: ConnectionFailure | null;
}
export interface ConnectionInstance {
  credential_slots: string[];
  runs: { operation_id: string; kind: string; outcome: string; message: string; at: string }[];
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
  resources: WasteSchedule[];
  last_error: ConnectionFailure | null;
  updated_at: string;
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
  mode: 'form' | 'yaml';
  dirty: boolean;
  scroll: number;
  focus?: string;
  document?: ProviderDefinition;
}
