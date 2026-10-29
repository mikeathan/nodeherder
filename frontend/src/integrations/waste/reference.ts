/** Stable source identity for a complete bin schedule. */
export interface WasteSummaryReference {
  kind: 'waste_summary';
  instance_id: string;
  resource_id: string;
}

/** Stable source identity for an individual saved bin service. */
export interface WasteServiceReference {
  kind: 'waste_service';
  instance_id: string;
  resource_id: string;
  service_id: string;
}

/** Includes persisted bin service references from earlier releases. */
export type WasteReference = WasteSummaryReference | WasteServiceReference;
