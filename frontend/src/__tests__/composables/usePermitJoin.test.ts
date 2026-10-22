import { effectScope, nextTick } from 'vue';
import { store } from '@/store';
import { usePermitJoin, usePermitJoinTracking } from '@/composables/usePermitJoin';

function report(permitJoin: boolean) {
  store.commit('hub/setBridgeSettings', { permitJoin, maxTimeAllowed: { value: 120, unit: 'seconds' } });
}

test('shell owns one countdown across consumer disposal and subsequent permit sessions', async () => {
  jest.useFakeTimers();
  report(false);
  const shell = effectScope();
  shell.run(usePermitJoinTracking);
  const firstConsumer = effectScope();
  const first = firstConsumer.run(usePermitJoin)!;
  const second = usePermitJoin();
  report(true); await nextTick();
  expect(first.remaining.value).toBe(120);
  expect(jest.getTimerCount()).toBe(1);
  jest.advanceTimersByTime(60000);
  expect(first.remaining.value).toBe(60);
  firstConsumer.stop();
  expect(second.remaining.value).toBe(60);
  report(false); await nextTick();
  expect(jest.getTimerCount()).toBe(0);
  report(true); await nextTick();
  expect(second.remaining.value).toBe(120);
  shell.stop();
  expect(jest.getTimerCount()).toBe(0);
  report(false);
  const remount = effectScope();
  remount.run(usePermitJoinTracking);
  report(true); await nextTick();
  expect(usePermitJoin().remaining.value).toBe(120);
  remount.stop();
  jest.useRealTimers();
});
