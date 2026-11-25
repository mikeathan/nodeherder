import type { Automation, Automations } from '@/types/automation.type';
import { downloadTextFile } from '@/services/logs';

/** Versioned client export envelope; import is intentionally not added in this branch. */
export type AutomationExport = {
  version: 1;
  exportedAt: string;
  automations: Automation[];
};

/** Serializes automations without depending on Vuex or transport details. */
export function serializeAutomations(automations: readonly Automation[], exportedAt = new Date()): string {
  const envelope: AutomationExport = {
    version: 1,
    exportedAt: exportedAt.toISOString(),
    automations: [...automations],
  };
  return JSON.stringify(envelope, null, 2);
}

/** Downloads one automation or the full set as a versioned JSON file. */
export function downloadAutomations(automations: readonly Automation[], name = 'nodeherder-automations'): void {
  const safeName = name.replace(/[^a-z0-9-_]+/gi, '-').toLowerCase();
  downloadTextFile(`${safeName}.json`, serializeAutomations(automations), 'application/json;charset=utf-8');
}

/** Convenience overload for callers that already hold the store list type. */
export function downloadAllAutomations(automations: Automations): void {
  downloadAutomations(automations);
}
