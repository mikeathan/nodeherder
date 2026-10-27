import { downloadHomeExport, readHomeExport } from '@/services/home-export.service';
import type { HomeLayout } from '@/types/home.type';

const layout: HomeLayout = { version: 1, revision: 1, etag: 'accepted', blocks: [] };
afterEach(() => jest.restoreAllMocks());

test('oversized exports are rejected before reading; malformed JSON never becomes a proposal', async () => {
  const text = jest.fn();
  const oversized = { size: 1_048_577, text } as unknown as File;
  await expect(readHomeExport(oversized, layout)).rejects.toThrow('smaller than 1 MiB');
  expect(text).not.toHaveBeenCalled();
  await expect(readHomeExport(new File(['{broken'], 'layout.json'), layout)).rejects.toThrow('not valid JSON');
});

test('download URL is released even if the browser cannot start the download', () => {
  jest.spyOn(URL, 'createObjectURL').mockReturnValue('blob:home-layout');
  const revoke = jest.spyOn(URL, 'revokeObjectURL').mockImplementation(() => undefined);
  const click = jest.fn(() => {
    throw new Error('download blocked');
  });
  jest.spyOn(document, 'createElement').mockReturnValue({ click } as unknown as HTMLAnchorElement);
  expect(() => downloadHomeExport(layout)).toThrow('download blocked');
  expect(revoke).toHaveBeenCalledWith('blob:home-layout');
});
