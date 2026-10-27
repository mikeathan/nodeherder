import type { HomeLayout } from '@/types/home.type';
import { parseHomeExport, serializeHomeExport } from '@/contracts/home-export';
import { HOME_LAYOUT_LIMITS } from '@/domain/home-layout';

/** Read and validate a bounded local export. The caller owns confirmation, cancellation and saving. */
export async function readHomeExport(file: File, current: HomeLayout): Promise<HomeLayout> {
  if (file.size > HOME_LAYOUT_LIMITS.fileBytes) throw new Error('Use a Home export smaller than 1 MiB.');
  let raw: unknown;
  try {
    raw = JSON.parse(await file.text());
  } catch {
    throw new Error('That file is not valid JSON. Choose a NodeHerder Home or areas export.');
  }
  return parseHomeExport(raw, current);
}

/** Download a reference-only snapshot, releasing its object URL even if the browser action fails. */
export function downloadHomeExport(layout: HomeLayout): void {
  const contents = serializeHomeExport(layout);
  const url = URL.createObjectURL(new Blob([contents], { type: 'application/json' }));
  try {
    const link = document.createElement('a');
    link.href = url;
    link.download = 'home-layout.json';
    link.click();
  } finally {
    URL.revokeObjectURL(url);
  }
}
