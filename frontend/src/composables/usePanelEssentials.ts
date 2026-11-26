/*
 * One browser-local preference for panel night essentials. Shared so a panel reload and the
 * picker observe the same list; storage failures keep the in-memory choice for this session.
 */
import { ref } from 'vue';
import { parsePanelEssentials, type PanelEssential } from '@/domain/panel-essentials';
import { safeStorage } from '@/utils/storage';

const STORAGE_KEY = 'nodeherder_panel_essentials';

function read(): PanelEssential[] {
  const raw = safeStorage.getItem(STORAGE_KEY);
  if (!raw) return [];
  try {
    return parsePanelEssentials(JSON.parse(raw));
  } catch {
    return [];
  }
}

const essentials = ref<PanelEssential[]>(read());

export function usePanelEssentials() {
  return {
    essentials,
    save(next: PanelEssential[]) {
      essentials.value = [...next];
      safeStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    },
  };
}
