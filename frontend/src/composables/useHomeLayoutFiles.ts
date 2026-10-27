import { onBeforeUnmount, readonly, ref } from 'vue';
import type { HomeLayoutScope } from './useHomeLayout';
import { confirm } from './useConfirm';
import { cloneHome } from '@/domain/home-layout';
import { downloadHomeExport, readHomeExport } from '@/services/home-export.service';

/** Own file-transfer interactions for one Home page. File IO/ingress belongs to the service;
 * this scope owns confirmation and prevents abandoned reads/dialogs from saving after navigation. */
export function useHomeLayoutFiles(home: Pick<HomeLayoutScope, 'layout' | 'save' | 'isActive'>) {
  const error = ref<string | null>(null);
  const importing = ref(false);
  let disposed = false;
  onBeforeUnmount(() => {
    disposed = true;
  });

  /** Download the currently accepted reference-only layout. */
  function exportLayout() {
    error.value = null;
    try {
      downloadHomeExport(home.layout.value);
    } catch {
      error.value = 'Home could not be exported. Try again.';
    }
  }

  /** Validate and confirm a local file against a captured revision; stale imports conflict rather
   * than being retagged onto a layout changed while the file/dialog was open. */
  async function importFile(file: File) {
    if (disposed || !home.isActive() || importing.value) return;
    importing.value = true;
    error.value = null;
    const current = cloneHome(home.layout.value);
    try {
      const proposal = await readHomeExport(file, current);
      if (disposed || !home.isActive()) return;
      const accepted = await confirm({
        title: 'Import Home layout?',
        message: `Replace Home placements using ${file.name}? Devices and integration configuration are kept.`,
        confirmLabel: 'Import',
      });
      if (accepted && !disposed && home.isActive()) await home.save(proposal);
    } catch (cause) {
      if (!disposed && home.isActive())
        error.value = cause instanceof Error ? cause.message : 'That Home export is invalid.';
    } finally {
      importing.value = false;
    }
  }

  return { error: readonly(error), importing: readonly(importing), exportLayout, importFile };
}
