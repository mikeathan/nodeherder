import { useConnections } from '@/composables/useConnections';
import { ConnectionError, createConnectionsApi } from '@/services/integrations.service';
import { createOperationObserver } from '@/domain/connection-operation-observer';

jest.mock('vue', () => ({ ...jest.requireActual('vue'), onBeforeUnmount: jest.fn() }));
jest.mock('vue-router', () => ({ useRouter: () => ({ replace: jest.fn() }) }));
jest.mock('@/store', () => ({ store: { state: { integrations: { generation: 1 } }, watch: () => jest.fn(), commit: jest.fn() } }));
jest.mock('@/services/integrations.service', () => {
  const actual = jest.requireActual('@/services/integrations.service');
  return { ...actual, createConnectionsApi: jest.fn() };
});
jest.mock('@/domain/connection-operation-observer', () => ({ createOperationObserver: jest.fn() }));

const check = jest.fn();
const start = jest.fn();
const stop = jest.fn();

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(createConnectionsApi).mockReturnValue({ check } as unknown as ReturnType<typeof createConnectionsApi>);
  jest.mocked(createOperationObserver).mockReturnValue({ start, stop });
});

test('checks the current operation with its refreshed revision and observes it', async () => {
  check.mockResolvedValue({ id: 'operation' });
  const reload = jest.fn().mockResolvedValue({ operation_id: 'operation', etag: 'new-revision' });
  await useConnections().checkOperation('operation', reload);
  expect(check).toHaveBeenCalledWith('operation', 'new-revision');
  expect(start).toHaveBeenCalledWith('operation');
});

test('does not check an operation replaced during the reload', async () => {
  const reload = jest.fn().mockResolvedValue({ operation_id: 'other', etag: 'revision' });
  await useConnections().checkOperation('operation', reload);
  expect(check).not.toHaveBeenCalled();
  expect(start).not.toHaveBeenCalled();
});

test('reloads after a revision conflict without submitting another check', async () => {
  check.mockRejectedValue(new ConnectionError({ code: 'revision_conflict', message: 'Changed' }, 409));
  const reload = jest.fn().mockResolvedValue({ operation_id: 'operation', etag: 'revision' });
  const scope = useConnections();
  await scope.checkOperation('operation', reload);
  expect(check).toHaveBeenCalledTimes(1);
  expect(reload).toHaveBeenCalledTimes(2);
  expect(scope.error.value).toBeNull();
  expect(start).not.toHaveBeenCalled();
});
