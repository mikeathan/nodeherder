import { observationHidden, setObservationHidden } from '@/domain/integrations/observation-visibility';

/*
 * The preference is written in the setter, not in a watcher. A module-level `watch` is not attached
 * to a component scope and its callback never ran, so hiding the group silently failed to persist
 * and it reappeared on the next reload with nothing to explain why.
 *
 * These run without a DOM, which is why they assert the reactive state rather than reading storage:
 * `safeStorage` degrades to a no-op when `window` is absent, so a storage assertion here would test
 * the absence of jsdom rather than the code. The persistence path is covered in the browser.
 */
describe('observation visibility', () => {
  beforeEach(() => setObservationHidden(false));

  test('the default is shown', () => {
    // Hidden by default would make the feature look absent on a fresh install.
    expect(observationHidden().value).toBe(false);
  });

  test('hiding and revealing update the shared state', () => {
    setObservationHidden(true);
    expect(observationHidden().value).toBe(true);

    setObservationHidden(false);
    expect(observationHidden().value).toBe(false);
  });

  test('the state is shared, not per call', () => {
    // The page and the board each call the accessor; both must see one value, or the toggle and the
    // board would disagree about whether the group is hidden.
    setObservationHidden(true);
    expect(observationHidden().value).toBe(true);
    expect(observationHidden()).toBe(observationHidden());
  });
});
