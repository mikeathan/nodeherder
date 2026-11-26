/*
 * Shared browser preference for the optional panel ambient-light trigger. Storage failures
 * keep the in-memory choice for the session.
 */
import { ref } from 'vue';
import { parsePanelAmbient, type PanelAmbientLight } from '@/domain/panel-ambient';
import { safeStorage } from '@/utils/storage';

const STORAGE_KEY = 'nodeherder_panel_ambient';

function read(): PanelAmbientLight | null {
  const raw = safeStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return parsePanelAmbient(JSON.parse(raw));
  } catch {
    return null;
  }
}

const ambient = ref<PanelAmbientLight | null>(read());

export function usePanelAmbient() {
  return {
    ambient,
    save(next: PanelAmbientLight | null) {
      ambient.value = next;
      if (next) safeStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      else safeStorage.removeItem(STORAGE_KEY);
    },
  };
}
