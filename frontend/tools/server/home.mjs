import { createHash } from 'node:crypto';

const clone = (value) => JSON.parse(JSON.stringify(value));

function layoutETag(layout) {
  const hash = createHash('sha256')
    .update(JSON.stringify({ ...layout, etag: '' }))
    .digest('hex');
  return `"${hash}"`;
}

function legacyFixtureBlock(group, groupIndex) {
  const cards = [];
  for (const [deviceIndex, device] of Object.values(group.deviceGroup ?? {}).entries()) {
    for (const [exposeIndex, expose] of device.exposes.entries()) {
      cards.push({
        id: `mock-card-${groupIndex}-${deviceIndex}-${exposeIndex}`,
        kind: 'device',
        panel: true,
        device_id: device.deviceId,
        expose,
      });
    }
  }
  return { id: `mock-group-${groupIndex}`, kind: 'group', name: group.name, cards };
}

function migrateFixtureGroups(groups) {
  const ordered = Object.values(groups).sort(
    (left, right) => (left.order ?? Infinity) - (right.order ?? Infinity) || left.name.localeCompare(right.name)
  );
  const layout = { version: 1, revision: 0, etag: '', blocks: ordered.map(legacyFixtureBlock) };
  layout.etag = layoutETag(layout);
  return layout;
}

function validFixtureLayout(layout) {
  if (layout?.version !== 1 || !Array.isArray(layout.blocks) || layout.blocks.length > 50) return false;
  const blockIds = new Set();
  const cardIds = new Set();
  for (const block of layout.blocks) {
    if (!block?.id || blockIds.has(block.id) || !['group', 'card'].includes(block.kind) || !Array.isArray(block.cards))
      return false;
    if (block.kind === 'card' && block.cards.length !== 1) return false;
    blockIds.add(block.id);
    for (const card of block.cards) {
      if (!card?.id || cardIds.has(card.id) || card.kind !== 'device') return false;
      cardIds.add(card.id);
      if (cardIds.size > 500) return false;
    }
  }
  return true;
}

function deviceAreaProjection(layout) {
  const groups = Object.create(null);
  for (const block of layout.blocks) {
    if (block.kind !== 'group') continue;
    const group = { name: block.name, order: Object.keys(groups).length, deviceGroup: Object.create(null) };
    for (const card of block.cards) {
      const device = group.deviceGroup[card.device_id] ?? { deviceId: card.device_id, exposes: [] };
      if (!device.exposes.includes(card.expose)) device.exposes.push(card.expose);
      group.deviceGroup[card.device_id] = device;
    }
    groups[block.name] = group;
  }
  return groups;
}

function fail(response, status, code, message) {
  return response.status(status).json({ data: null, error: { code, message } });
}

/** Minimal device-only fixture endpoint. Integration/source/transaction acceptance is tested
 * against the real isolated Go hub, so this mock never becomes a second production owner. */
export function registerHomeRoutes(app, state) {
  let accepted = null;
  function current() {
    return accepted ? clone(accepted) : migrateFixtureGroups(state().config.hub.dashboardGroups);
  }

  app.get('/api/home/layout', (request, response) => {
    response.set('Cache-Control', 'private, no-store');
    if (!request.cookies?.session) return fail(response, 401, 'unauthorized', 'Sign in to use Home.');
    const data = current();
    return response.set('ETag', data.etag).json({ data, error: null });
  });

  app.put('/api/home/layout', (request, response) => {
    response.set('Cache-Control', 'private, no-store');
    if (!request.cookies?.session) return fail(response, 401, 'unauthorized', 'Sign in to use Home.');
    const trustedOrigin = request.get('Origin') === 'http://localhost:4100';
    const requestMarker = request.get('X-NodeHerder-Request') === '1';
    if (!trustedOrigin || !requestMarker) return fail(response, 403, 'csrf_rejected', 'Use the configured fixture UI.');
    const previous = current();
    const expected = request.get('If-Match');
    if (!expected) return fail(response, 428, 'precondition_required', 'Reload Home.');
    if (expected !== previous.etag) return fail(response, 409, 'revision_conflict', 'Home changed. Reload the layout.');
    if (!validFixtureLayout(request.body)) return fail(response, 400, 'invalid_input', 'Invalid fixture layout.');
    accepted = clone({ ...request.body, revision: previous.revision + 1, etag: '' });
    accepted.etag = layoutETag(accepted);
    state().config.hub.dashboardGroups = deviceAreaProjection(accepted);
    return response.set('ETag', accepted.etag).json({ data: clone(accepted), error: null });
  });
}
