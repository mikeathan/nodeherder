import { test, expect, signIn } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';
import { ESPHOME_DEVICES_PATH } from '../src/contracts/esphome';

// Spec 007 US-03: device list and device page.
test.describe('devices', () => {
  test('device list shows every hub device and opens a device page', async ({ page }) => {
    await signIn(page, ui.devices.listPath);
    await expect(ui.devices.listRows(page)).toHaveCount(MOCK.deviceCount);
    await ui.devices.openFromList(page, MOCK.atticSocket.name);
    await expect(page).toHaveURL(new RegExp(MOCK.atticSocket.id));
    await expect(ui.devices.pageTitle(page)).toContainText(MOCK.atticSocket.name);
  });

  test('ESPHome enrollment and management use the existing device flow', async ({ page }) => {
    const requests: { method: string; body?: Record<string, unknown> }[] = [];
    let item: Record<string, unknown> | null = null;
    await page.route(`**/api${ESPHOME_DEVICES_PATH}**`, async (route) => {
      const request = route.request();
      const method = request.method();
      const body = ['POST', 'PUT'].includes(method) ? request.postDataJSON() as Record<string, unknown> : undefined;
      requests.push({ method, body });
      if (method === 'POST') item = { id: 'fixture-id', provider: 'esphome', name: body?.name, address: body?.address, port: body?.port, status: 'offline', diagnostic: 'connection_failed' };
      if (method === 'PUT' && item) item = { ...item, ...body };
      if (method === 'DELETE') item = null;
      await route.fulfill({ status: method === 'POST' ? 201 : 200, json: { data: method === 'GET' ? (item ? [item] : []) : item ?? { removed: true }, error: null } });
    });
    await signIn(page, '/devicedashboard');
    await page.getByRole('button', { name: 'Add ESPHome device' }).click();
    const add = page.getByRole('dialog', { name: 'Add ESPHome device' });
    await add.getByLabel('Name').fill('Radar');
    await add.getByLabel('IP address').fill('192.168.1.2');
    await add.getByLabel('Native API port').fill('6053');
    await add.getByLabel('Encryption key (optional)').fill('test-key');
    await add.getByRole('button', { name: 'Add device' }).click();
    await expect(add).toBeHidden();
    expect(requests.find((r) => r.method === 'POST')?.body).toMatchObject({ address: '192.168.1.2', port: 6053, encryption_key: 'test-key' });
    const manage = page.getByRole('dialog', { name: 'Manage ESPHome devices' });
    await expect(manage).toBeVisible();
    await expect(manage).toContainText('offline');
    await expect(manage).toContainText('Check the address');
    await manage.getByRole('button', { name: 'Edit Radar' }).click();
    await manage.getByLabel('IP address').fill('192.168.1.3');
    await manage.getByRole('button', { name: 'Save connection' }).click();
    await expect(manage).toContainText('192.168.1.3');
    expect(requests.find((r) => r.method === 'PUT')?.body).not.toHaveProperty('encryption_key');
    await manage.getByRole('button', { name: 'Remove Radar' }).click();
    await page.getByRole('dialog', { name: 'Remove Radar?' }).getByRole('button', { name: 'Remove device' }).click();
    await expect(manage).toContainText('No ESPHome devices have been added.');
    expect(requests.some((r) => r.method === 'DELETE')).toBe(true);
  });

  test('ESPHome management distinguishes missing connection settings from an empty list', async ({ page }) => {
    await page.route('**/api/hubstate', async (route) => {
      const response = await route.fetch();
      const state = await response.json();
      const device = state.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
      device.provider = 'esphome';
      device.connection_type = 'esphome';
      await route.fulfill({ response, json: state });
    });
    let fail = false;
    await page.route(`**/api${ESPHOME_DEVICES_PATH}`, (route) => fail
      ? route.fulfill({ status: 503, json: { data: null, error: { code: 'unavailable' } } })
      : route.fulfill({ json: { data: [], error: null } }));
    await signIn(page, '/devicedashboard');
    await page.getByRole('button', { name: 'Manage ESPHome' }).click();
    const dialog = page.getByRole('dialog', { name: 'Manage ESPHome devices' });
    await expect(dialog).toContainText('no matching saved connection settings');
    await expect(dialog).toContainText(MOCK.livingRoomLight.name);
    await expect(dialog.getByText('No ESPHome devices have been added.')).toHaveCount(0);
    await dialog.getByRole('button', { name: 'Close', exact: true }).last().click();
    fail = true;
    await page.getByRole('button', { name: 'Manage ESPHome' }).click();
    await expect(dialog.getByRole('alert')).toContainText('Unable to load ESPHome settings');
    await expect(dialog.getByText('No ESPHome devices have been added.')).toHaveCount(0);
  });

  test('reopening ESPHome management ignores the previous request', async ({ page }) => {
    let calls = 0;
    let releaseFirst!: () => void;
    let finishFirst!: () => void;
    const firstBlocked = new Promise<void>((resolve) => { releaseFirst = resolve; });
    const firstFinished = new Promise<void>((resolve) => { finishFirst = resolve; });
    await page.route(`**/api${ESPHOME_DEVICES_PATH}`, async (route) => {
      calls++;
      if (calls === 1) {
        await firstBlocked;
        try { await route.fulfill({ json: { data: [], error: null } }); } catch { /* the dialog canceled this request */ }
        finally { finishFirst(); }
        return;
      }
      await route.fulfill({ json: { data: [{ id: 'fixture', name: 'Radar', address: '192.168.1.2', port: 6053, status: 'online' }], error: null } });
    });
    try {
      await signIn(page, '/devicedashboard');
      const open = page.getByRole('button', { name: 'Manage ESPHome' });
      await open.click();
      const dialog = page.getByRole('dialog', { name: 'Manage ESPHome devices' });
      await expect(dialog.getByText('Loading devices…')).toBeVisible();
      await dialog.getByRole('button', { name: 'Close', exact: true }).last().click();
      await open.click();
      await expect(dialog.getByText('192.168.1.2:6053')).toBeVisible();
      releaseFirst();
      await firstFinished;
      await expect(dialog.getByText('192.168.1.2:6053')).toBeVisible();
      await expect(dialog.getByText('No ESPHome devices have been added.')).toHaveCount(0);
    } finally { releaseFirst(); }
  });

  test('ESPHome entity sections are sorted, searchable and paginated', async ({ page }, info) => {
    await page.route('**/api/hubstate', async (route) => {
      const response = await route.fetch();
      const state = await response.json();
      const device = state.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
      device.provider = 'esphome';
      device.connection_type = 'esphome';
      device.availability = 'online';
      device.exposes = {};
      for (const category of ['measurement', 'config', 'diagnostic']) {
        for (let index = 12; index >= 1; index--) {
          const name = `${category} ${String(index).padStart(2, '0')}`;
          device.exposes[`${category}_${index}`] = { name: `${category}_${index}`, description: name, unit: '', data: index, type: 'numeric', category, access_mode: 'read', attributes: { label: name }, values: {} };
        }
      }
      await route.fulfill({ response, json: state });
    });
    let sendUpdate: ((payload: unknown) => void) | undefined;
    await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
      sendUpdate = (payload) => socket.send(JSON.stringify({ type: 'deviceUpdated', payload }));
      socket.onMessage(() => {});
    });
    await signIn(page, `/devicepage/${MOCK.livingRoomLight.id}`);
    const configuration = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Device configuration', exact: true }) });
    const readings = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Readings', exact: true }) });
    const diagnostics = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Diagnostics', exact: true }) });
    await page.setViewportSize({ width: 1440, height: 900 });
    const [readingsBox, diagnosticsBox, configurationBox] = await Promise.all([readings.boundingBox(), diagnostics.boundingBox(), configuration.boundingBox()]);
    expect(readingsBox).not.toBeNull();
    expect(diagnosticsBox).not.toBeNull();
    expect(configurationBox).not.toBeNull();
    expect(Math.abs(diagnosticsBox!.x - readingsBox!.x)).toBeLessThan(1);
    expect(diagnosticsBox!.y).toBeGreaterThan(readingsBox!.y);
    expect(configurationBox!.x).toBeGreaterThan(readingsBox!.x);
    await expect(configuration.getByText('config 01', { exact: true })).toBeVisible();
    await expect(configuration.getByText('config 09', { exact: true })).toHaveCount(0);
    await configuration.getByRole('button', { name: 'Next device configuration page' }).click();
    await expect(configuration.getByText('config 09', { exact: true })).toBeVisible();
    await expect.poll(() => !!sendUpdate).toBe(true);
    sendUpdate!({ id: MOCK.livingRoomLight.id, data: { config_9: 100 } });
    await expect(configuration.getByText('100', { exact: true })).toBeVisible();
    await expect(configuration.getByText('config 09', { exact: true })).toBeVisible();
    await configuration.getByLabel('Search device configuration').fill('12');
    await expect(configuration.getByText('config 12', { exact: true })).toBeVisible();
    await expect(configuration.getByText('config 09', { exact: true })).toHaveCount(0);
    for (const width of info.project.name === 'desktop' ? [1440, 768, 375] : [page.viewportSize()!.width]) {
      await page.setViewportSize({ width, height: 900 });
      expect(await page.locator('html').evaluate((html) => html.scrollWidth <= html.clientWidth)).toBe(true);
    }
  });

  test('spatial mappings save and show only current provider reports', async ({ page }, info) => {
    await page.route('**/api/hubstate', async (route) => {
      const response = await route.fetch();
      const state = await response.json();
      const device = state.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
      device.provider = 'esphome';
      device.availability = 'online';
      const entity = (name: string, type: string, unit: string, data: unknown) => ({ name, description: '', unit, data, type, category: 'measurement', access_mode: 'read', attributes: { label: name }, values: {} });
      device.exposes.target_x = entity('target_x', 'numeric', 'mm', 100);
      device.exposes.target_y = entity('target_y', 'numeric', 'mm', 200);
      device.exposes.target2_x = entity('target2_x', 'numeric', 'mm', 300);
      device.exposes.target2_y = entity('target2_y', 'numeric', 'mm', 400);
      device.exposes.target2_active = entity('target2_active', 'binary', '', true);
      device.exposes.target3_x = entity('target3_x', 'numeric', 'mm', 300);
      device.exposes.target_active = entity('target_active', 'binary', '', true);
      device.exposes.target3_active = entity('target3_active', 'binary', '', true);
      for (const name of ['beginX','beginY','endX','endY']) device.exposes[name]={...entity(name,'numeric','mm',0),access_mode:'readwrite',category:'config',attributes:{label:name,min:-6000,max:6000,step:10}};
      device.exposes.target_x.attributes.label = 'Target 1 X';
      device.exposes.target_y.attributes.label = 'Target 1 Y';
      device.exposes.target2_x.attributes.label = 'Target 2 X';
      device.exposes.target2_y.attributes.label = 'Target 2 Y';
      device.exposes.target2_active.attributes.label = 'Target 2 Active';
      device.exposes.target3_x.attributes.label = 'Target 3 X';
      device.exposes.target_active.attributes.label = 'Target 3 Active';
      device.exposes.target3_active.attributes.label = 'Target 1 Active';
      // Start from a saved complete Target 2 mapping while live reports continue.
      const overrides = state.config.hub.devices.overrides ?? (state.config.hub.devices.overrides = {});
      overrides[MOCK.livingRoomLight.id] = {
        ...(overrides[MOCK.livingRoomLight.id] ?? {}),
        id: MOCK.livingRoomLight.id,
        disabled: false,
        metricsEnabled: false,
        spatial: { mappings: [{x:'',y:'',active:''},{x:'target2_x',y:'target2_y',active:'target2_active'},{x:'',y:'',active:''}], scale: 6000, orientation: 0 },
      };
      // A hub-derived presence device (spec 041): the tab must show its values as the
      // authority instead of pretending the geometric overlay is the server truth.
      state.devices.push({
        id: `presence-${MOCK.livingRoomLight.id}`,
        provider: 'presence',
        connection_type: 'derived',
        friendly_name: 'Living Room Light (presence)',
        description: 'Derived presence',
        power_source: 'mains',
        availability: 'online',
        last_seen: new Date().toISOString(),
        properties: { source_device_id: MOCK.livingRoomLight.id, derived: true },
        exposes: {
          occupied: entity('occupied', 'binary', '', true),
          target_count: entity('target_count', 'numeric', '', 1),
          last_zone: entity('last_zone', 'text', '', 'Desk'),
          empty_seconds: entity('empty_seconds', 'numeric', 's', 0),
        },
      });
      await route.fulfill({ response, json: state });
    });
    let sendUpdate: ((payload: unknown) => void) | undefined;
    let sendSpatial: ((payload: unknown) => void) | undefined;
    const zoneCommands: {name:string;value:number}[]=[];
    let rejectSpatialSave=false;
    let savedMapping: unknown;
    let savedMappings: unknown;
    let savedWorkspace: {zones:{name:string;points:{x:number;y:number}[]}[];objects:{name:string}[];outline:unknown[]} | undefined;
    await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
      sendUpdate = (payload) => socket.send(JSON.stringify({ type: 'deviceUpdated', payload }));
      sendSpatial = (payload) => socket.send(JSON.stringify({ type: 'spatialTargets', payload }));
      socket.onMessage((raw) => {
        const frame = JSON.parse(String(raw));
        if (frame.type==='deviceSetValue') zoneCommands.push(frame.payload);
        if (frame.type === 'saveDeviceConfigOverride') {
          savedMappings = frame.payload.spatial?.mappings;
          savedMapping = frame.payload.spatial?.mappings?.[0];
          savedWorkspace = frame.payload.spatial?.workspace;
          socket.send(JSON.stringify({
            type: 'deviceConfigSaveResult',
            payload: {
              requestId: frame.payload.requestId,
              deviceId: frame.payload.id,
              success: !rejectSpatialSave,
              ...(rejectSpatialSave ? { error: 'Spatial save rejected' } : {}),
            },
          }));
        }
      });
    });
    await signIn(page, `/devicepage/${MOCK.livingRoomLight.id}`);
    // Presence is a first-class tab for mappable sensors. Its header chip is the entry point,
    // and the Controls tab no longer draws a second radar.
    const presenceTab = page.getByRole('tab', { name: 'Presence', exact: true });
    await expect(presenceTab).toBeVisible();
    const presenceChip = page.getByRole('button', { name: /^Presence: .*Open the presence view$/ });
    await expect(presenceChip).toBeVisible();
    await expect(page.getByRole('img', { name: /Live room canvas/ })).toHaveCount(0);
    await presenceChip.click();
    await expect(presenceTab).toHaveAttribute('aria-selected', 'true');
    await expect(page.getByRole('heading', { name: 'Presence', exact: true })).toBeVisible();
    await expect(page.getByRole('img', { name: /Live room canvas/ })).toHaveCount(1);
    await expect(page.getByText(/Server presence: occupied · 1 target · last zone Desk/)).toBeVisible();
    // The shared view panel groups the display controls on one baseline.
    if (page.viewportSize()!.width >= 900) {
      const viewControls = page.getByRole('button', { name: 'View controls' });
      if ((await viewControls.getAttribute('aria-expanded')) !== 'true') {
        await viewControls.click();
      }
      const selects = await page
        .locator('.nh-view-body select')
        .evaluateAll((nodes) =>
          nodes.map((node) => {
            const box = node.getBoundingClientRect();
            return { bottom: box.bottom };
          })
        );
      expect(selects.length).toBeGreaterThanOrEqual(3);
      const bottoms = selects.map((select) => select.bottom);
      expect(Math.max(...bottoms) - Math.min(...bottoms)).toBeLessThanOrEqual(1);
    }
    const help = page.locator('summary').filter({ hasText: 'Help with the presence view' });
    await help.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('heading', { name: 'Map a target' })).toBeVisible();
    await expect(page.getByRole('img', { name: /Example coordinate plane/ })).toBeVisible();
    await help.press('Enter');
    await expect(page.getByRole('heading', { name: 'Map a target' })).toHaveCount(0);
    await page.locator('summary').filter({ hasText: /^Target mapping$/ }).click();
    // Live coordinate reports must not rebuild an open mapping selector.
    await expect.poll(() => !!sendUpdate).toBe(true);
    sendUpdate!({ id: MOCK.livingRoomLight.id, data: { target_x: 101, target2_x: 301 } });
    // Every row must accept and clear its own selection, not only Target 1.
    const target2 = page.getByRole('group', { name: 'Target 2' });
    for (const [role, value] of [['X','target2_x'],['Y','target2_y'],['Active','target2_active']] as const) {
      await target2.getByRole('combobox',{name:role,exact:true}).selectOption(value);
      await expect(target2.getByRole('combobox',{name:role,exact:true})).toHaveValue(value);
    }
    // Native dropdowns can preview a choice before committing a change event.
    const pendingSelect = target2.getByRole('combobox', {name:'X',exact:true});
    await pendingSelect.evaluate((select: HTMLSelectElement) => {
      select.focus();
      select.value = '';
      select.dispatchEvent(new Event('input', {bubbles:true}));
    });
    sendUpdate!({ id: MOCK.livingRoomLight.id, data: { target_x: 102, target2_x: 302 } });
    await page.waitForTimeout(100);
    await expect(pendingSelect).toHaveValue('');
    await pendingSelect.dispatchEvent('change');
    for (const role of ['X','Y','Active'] as const) {
      const select = target2.getByRole('combobox', { name: role, exact: true });
      await select.focus();
      await select.press('n'); // native typeahead selects Not mapped
      await select.press('Tab');
      await expect(select).toHaveValue('');
    }
    const target = page.getByRole('group', { name: 'Target 1' });
    await expect(target.getByRole('combobox', { name: 'X', exact: true }).locator('option[value="target3_x"]')).toHaveCount(0);
    await target.getByRole('combobox', { name: 'X', exact: true }).selectOption('target_x');
    await page.getByRole('button', { name: 'Save spatial settings' }).click();
    await expect(page.getByRole('alert')).toContainText('Target 1: choose an X, Y and Active entity');
    expect(savedMapping).toBeUndefined();
    await expect.poll(() => !!sendUpdate).toBe(true);
    sendUpdate!({ id: MOCK.livingRoomLight.id, data: { target_x: 101 } });
    await expect(target.getByRole('combobox', { name: 'X', exact: true })).toHaveValue('target_x');
    await target.getByRole('combobox', { name: 'Y', exact: true }).selectOption('target_y');
    sendUpdate!({ id: MOCK.livingRoomLight.id, data: { target_y: 201 } });
    await expect(target.getByRole('combobox', { name: 'Y', exact: true })).toHaveValue('target_y');
    await expect(target.getByRole('combobox', { name: 'Active', exact: true }).locator('option[value="target_active"]')).toHaveCount(0);
    await target.getByRole('combobox', { name: 'Active', exact: true }).selectOption('target3_active');
    sendUpdate!({ id: MOCK.livingRoomLight.id, data: { target_x: 102 } });
    await expect(target.getByRole('combobox', { name: 'Active', exact: true })).toHaveValue('target3_active');
    await page.getByRole('button', { name: 'Save spatial settings' }).click();
    await expect.poll(() => savedMapping).toEqual({ x: 'target_x', y: 'target_y', active: 'target3_active' });
    await expect.poll(() => savedMappings).toEqual([
      { x: 'target_x', y: 'target_y', active: 'target3_active' },
      { x: '', y: '', active: '' },
      { x: '', y: '', active: '' },
    ]);
    await page.getByRole('tab', { name: 'Controls', exact: true }).click();
    await expect(page.getByRole('img', { name: /Live room canvas/ })).toHaveCount(0);
    await page.getByRole('tab', { name: 'Presence', exact: true }).click();
    await page.locator('summary').filter({ hasText: /^Target mapping$/ }).click();
    for (const role of ['X', 'Y', 'Active'] as const) {
      await expect(target2.getByRole('combobox', { name: role, exact: true })).toHaveValue('');
    }
    await expect(target.getByRole('combobox', { name: 'X', exact: true })).toHaveValue('target_x');
    sendUpdate!({ id: MOCK.livingRoomLight.id, data: { target_x: 100, target_y: 200, target3_active: true } });
    await expect(page.getByText(/Track 1 \(live\): X/)).toBeVisible();
    const widths = info.project.name === 'desktop' ? [1440, 768, 375] : [page.viewportSize()!.width];
    for (const width of widths) {
      await page.setViewportSize({ width, height: 900 });
      expect(await page.locator('html').evaluate((html) => html.scrollWidth <= html.clientWidth)).toBe(true);
      await page.screenshot({ path: info.outputPath(`spatial-view-${width}.png`), fullPage: true });
    }

    // A complete frame wins over sampled coordinates and is labelled as frames.
    await expect.poll(() => !!sendSpatial).toBe(true);
    sendSpatial!({
      id: MOCK.livingRoomLight.id,
      at: new Date().toISOString(),
      targets: [
        { id: 1, active: true, x: { value: 1000, unit: 'mm' }, y: { value: 2000, unit: 'mm' } },
        { id: 2, active: false },
      ],
    });
    await expect(page.getByText('Live frames')).toBeVisible();
    await expect(page.getByText(/Track 1 \(live\): X/)).toBeVisible();

    // View controls are grouped in the shared viewer panel; open it at narrow widths.
    const viewControls = page.getByRole('button', { name: 'View controls' });
    if ((await viewControls.getAttribute('aria-expanded')) !== 'true') {
      await viewControls.click();
    }
    const smoothing = page.getByRole('combobox', { name: 'Smoothing' });
    await smoothing.selectOption('off');
    await expect(smoothing).toHaveValue('off');
    await smoothing.selectOption('high');
    await expect(smoothing).toHaveValue('high');
    await page.getByLabel('Raw', { exact: true }).check();
    await expect(page.getByRole('img', { name: /Live room canvas/ })).toBeVisible();
    await page.getByLabel('Raw', { exact: true }).uncheck();

    // Distance labels: derived from the position actually drawn (hypot(1000, 2000) ≈ 2.2 m),
    // quantised to 0.1 m so filter noise cannot rewrite the text every frame.
    const distance = page.getByLabel('Distance', { exact: true });
    await expect(distance).not.toBeChecked();
    await expect(page.locator('text.nh-target-label').first()).toHaveText('Track 1');
    await distance.focus();
    await expect(distance).toBeFocused();
    await distance.check();
    await expect(distance).toBeChecked();
    await expect(page.locator('text.nh-target-label').first()).toHaveText(/^Track 1 · 2\.2 m$/);
    await distance.uncheck();
    await expect(page.locator('text.nh-target-label').first()).toHaveText('Track 1');

    // A captured slot survives ageing (the 250 ms tick must not reset the selection).
    await page.getByRole('button', { name: 'Setup', exact: true }).click();
    await page.getByRole('button', { name: 'Calibration', exact: true }).click();
    const liveSlot = page.getByRole('combobox', { name: 'Firmware slot (measured)' }).first();
    // Configured slots are always selectable, even before a target has reported, and their
    // labels stay static while live values update beside the control.
    await expect(liveSlot.locator('option')).toHaveCount(4);
    await expect(liveSlot.locator('option').nth(1)).toHaveText('Slot 1');
    await liveSlot.selectOption({ label: 'Slot 1' });
    await expect(liveSlot).toHaveValue('1');
    await expect(page.getByText(/Latest raw .* mm at/)).toBeVisible();
    await page.waitForTimeout(3500);
    await expect(liveSlot).toHaveValue('1');
    await expect(page.getByText(/Captured .* mm at/)).toBeVisible();

    await page.getByRole('button', { name: 'Live', exact: true }).click();

    await page.getByRole('button', {name:'Create your first zone',exact:true}).click();
    await page.getByRole('button', {name:'Add detection zone',exact:true}).click();
    await page.getByLabel('Zone name',{exact:true}).fill('Desk');
    await page.getByLabel('Zone name',{exact:true}).press('Tab');
    await page.getByLabel('Zone width (mm)',{exact:true}).fill('1500');
    await page.getByLabel('Zone width (mm)',{exact:true}).press('Tab');
    rejectSpatialSave=true;
    await page.getByRole('button', {name:'Save spatial settings',exact:true}).click();
    await expect(page.getByRole('alert')).toContainText('Spatial save rejected');
    await expect(page.getByLabel('Zone width (mm)',{exact:true})).toHaveValue('1500');
    await expect(page.getByRole('button',{name:'Undo last edit',exact:true})).toBeEnabled();
    rejectSpatialSave=false;
    await page.getByRole('button', {name:'Save spatial settings',exact:true}).click();
    await expect(page.getByText('Spatial settings saved in NodeHerder.',{exact:false})).toBeVisible();
    await expect.poll(()=>savedWorkspace?.zones[0].name).toBe('Desk');
    await expect.poll(()=>savedWorkspace?.zones[0].points[1].x).toBe(1000);
    await page.locator('summary').filter({hasText:'Map physical zone coordinates'}).click();
    for (const role of ['beginX','beginY','endX','endY']) await page.getByRole('combobox',{name:`${role} entity`,exact:true}).selectOption(role);
    await page.getByRole('button',{name:'Apply rectangle to device',exact:true}).click();
    await expect.poll(()=>zoneCommands.length).toBe(4);
    expect(zoneCommands.map(c=>c.value)).toEqual([-500,1000,1000,2000]);
    await expect(page.getByText('Coordinates requested.',{exact:false})).toBeVisible();
    await page.getByRole('button', {name:'Remove zone',exact:true}).click();
    await page.getByRole('button', {name:'Undo last edit',exact:true}).click();
    await expect(page.getByRole('combobox',{name:'Zone slots',exact:true}).getByRole('option',{name:'Desk · detection'})).toHaveCount(1);
    const viewerViewport = page.viewportSize()!;
    const appMain = page.locator('.nh-app-main');
    await page.setViewportSize({ width: 1280, height: 600 });
    await page.getByRole('tab',{name:'History',exact:true}).click();
    await page.getByRole('button',{name:'Positions',exact:true}).click();
    // The replay viewer keeps its own options: no Distance checkbox, since only the live
    // panel wires that value into its canvas.
    await expect(page.getByLabel('Distance', { exact: true })).toHaveCount(0);
    await expect(page.locator('.nh-session-row')).toHaveCount(3);
    const fovToggle = page.getByLabel('Field of view', { exact: true });
    await expect(fovToggle).toBeEnabled();
    await expect(fovToggle).not.toBeChecked();
    await fovToggle.check();
    await expect(page.locator('polygon.nh-fov')).toHaveCount(1);
    await fovToggle.uncheck();
    await expect(page.locator('polygon.nh-fov')).toHaveCount(0);
    await expect(page.locator('circle.nh-target').first()).toBeVisible();
    await expect(page.locator('polyline.nh-trail:not(.is-recent)')).toHaveCount(1);
    await page.getByLabel('Path', { exact: true }).uncheck();
    await expect(page.locator('polyline.nh-trail')).toHaveCount(0);
    await page.getByLabel('Path', { exact: true }).check();
    await page.getByRole('button',{name:'Play',exact:true}).click();
    await expect(page.getByRole('button',{name:'Pause',exact:true})).toBeVisible();
    await page.getByRole('button',{name:'Pause',exact:true}).click();

    // AC-17: tablet and desktop widths keep the player in flow without sideways overflow.
    await page.setViewportSize({ width: 1440, height: 900 });
    await expect(page.locator('.nh-transport')).toBeVisible();
    expect(await page.locator('html').evaluate((html) => html.scrollWidth <= html.clientWidth)).toBe(true);

    // AC-21: the recordings rail collapses and restores without changing the selection.
    await expect(page.locator('.nh-positions-rail')).toBeVisible();
    const mainBefore = await page.locator('.nh-positions-main').boundingBox();
    const canvasBefore = await page.locator('.nh-spatial-canvas').boundingBox();
    await page.getByRole('button', { name: 'Hide recordings', exact: true }).click();
    await expect(page.locator('.nh-positions-rail')).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Show recordings (3)', exact: true })).toBeVisible();
    await expect(page.locator('.nh-coverage-segment[data-session-id="mock-session-evening"]')).toHaveAttribute('aria-pressed', 'true');
    const canvasExpanded = await page.locator('.nh-spatial-canvas').boundingBox();
    expect(canvasExpanded!.width).toBeGreaterThan(canvasBefore!.width);
    expect(Math.abs(canvasExpanded!.height - canvasBefore!.height)).toBeLessThanOrEqual(1);
    await page.getByRole('button', { name: 'Show recordings (3)', exact: true }).click();
    await expect(page.locator('.nh-positions-rail')).toBeVisible();

    // AC-22: the viewer is wide and short, and zoom scales only its contents.
    const canvasRestored = await page.locator('.nh-spatial-canvas').boundingBox();
    expect(Math.abs(canvasRestored!.width - mainBefore!.width)).toBeLessThanOrEqual(2);
    expect(canvasRestored!.height).toBeGreaterThanOrEqual(16 * 16 - 1);
    expect(canvasRestored!.height).toBeLessThanOrEqual(24 * 16 + 1);
    const viewBoxWidth = async () =>
      Number((await page.locator('.nh-spatial-canvas').getAttribute('viewBox'))!.split(' ')[2]);
    const canvasBox = await page.locator('.nh-spatial-canvas').boundingBox();
    const currentView = (await page.locator('.nh-spatial-canvas').getAttribute('viewBox'))!.split(' ').map(Number);
    expect(Math.abs(currentView[2] / currentView[3] - canvasBox!.width / canvasBox!.height)).toBeLessThan(0.05);
    const wideView = await viewBoxWidth();
    await page.getByRole('button', { name: '2×', exact: true }).click();
    await expect(page.getByRole('button', { name: '2×', exact: true })).toHaveAttribute('aria-pressed', 'true');
    const zoomedView = await viewBoxWidth();
    expect(zoomedView).toBeLessThan(wideView * 0.7);
    // The pinned device stays visible at the edge and the active point stays in frame.
    await expect(page.locator('.nh-device-bg')).toBeVisible();
    await expect(page.locator('circle.nh-target').first()).toBeVisible();
    await page.getByRole('button', { name: '5×', exact: true }).click();
    await expect(page.getByRole('button', { name: '5×', exact: true })).toHaveAttribute('aria-pressed', 'true');
    await expect(page.locator('.nh-device-bg')).toBeVisible();
    await expect(page.locator('circle.nh-target').first()).toBeVisible();
    await page.getByRole('button', { name: '1×', exact: true }).click();
    await expect(page.getByRole('button', { name: '1×', exact: true })).toHaveAttribute('aria-pressed', 'true');

    // Dragging pans the reader; Reset view restores the fitted view.
    const fittedView = await page.locator('.nh-spatial-canvas').getAttribute('viewBox');
    const panBox = await page.locator('.nh-spatial-canvas').boundingBox();
    await page.mouse.move(panBox!.x + panBox!.width / 2, panBox!.y + panBox!.height / 2);
    await page.mouse.down();
    await page.mouse.move(panBox!.x + panBox!.width / 2 + 80, panBox!.y + panBox!.height / 2 + 30);
    await page.mouse.up();
    expect(await page.locator('.nh-spatial-canvas').getAttribute('viewBox')).not.toBe(fittedView);
    await page.getByRole('button', { name: 'Reset view', exact: true }).click();
    await expect.poll(async () => page.locator('.nh-spatial-canvas').getAttribute('viewBox')).toBe(fittedView);

    // The field-of-view cone is available on the heatmap canvas too.
    await page.getByRole('button', { name: 'Heatmap', exact: true }).click();
    await expect(page.getByRole('img', { name: 'Recorded heatmap canvas' })).toBeVisible();
    const heatFov = page.getByLabel('Field of view', { exact: true });
    await heatFov.check();
    await expect(page.locator('polygon.nh-fov')).toHaveCount(1);
    await heatFov.uncheck();
    await expect(page.locator('polygon.nh-fov')).toHaveCount(0);
    await page.getByRole('button', { name: 'Player', exact: true }).click();
    await expect(page.locator('.nh-coverage-segment[data-session-id="mock-session-evening"]')).toHaveAttribute('aria-pressed', 'true');

    await page.setViewportSize({ width: 768, height: 1024 });
    expect(await page.locator('html').evaluate((html) => html.scrollWidth <= html.clientWidth)).toBe(true);
    await page.setViewportSize({ width: 1280, height: 600 });

    // Spec 057 AC-01: choosing a recording re-renders in place and never moves the page.
    const range = page.getByLabel('Playback position');
    await range.scrollIntoViewIfNeeded();
    const beforeScroll = await appMain.evaluate((element) => element.scrollTop);
    expect(beforeScroll).toBeGreaterThan(0);
    await page.locator('.nh-session-row[data-session-id="mock-session-morning"]').click();
    await expect(page.locator('.nh-session-row[data-session-id="mock-session-morning"]')).toHaveAttribute('aria-pressed', 'true');
    await expect(page.locator('.nh-coverage-segment[data-session-id="mock-session-morning"]')).toHaveAttribute('aria-pressed', 'true');
    expect(Math.abs((await appMain.evaluate((element) => element.scrollTop)) - beforeScroll)).toBeLessThanOrEqual(2);

    // AC-10/AC-12: movement episodes skip the recorded pause and mark the jump.
    await expect(page.locator('line.nh-hop')).toHaveCount(0);
    await page.getByRole('button', { name: 'Next movement', exact: true }).click();
    await expect(page.locator('line.nh-hop')).toHaveCount(1);
    await expect(page.locator('circle.nh-pulse')).toHaveCount(1);

    // AC-11: a playhead inside the pause says what the missing samples mean.
    const sessionStart = Number(await range.getAttribute('min'));
    await range.evaluate((element, value) => {
      const input = element as HTMLInputElement;
      input.value = String(value);
      input.dispatchEvent(new Event('input', { bubbles: true }));
      input.dispatchEvent(new Event('change', { bubbles: true }));
    }, sessionStart + 100_000);
    await expect(page.getByText(/No movement recorded for/)).toBeVisible();

    // Playback must advance even when the playhead is inside a recorded gap.
    const idleTime = Number(await range.inputValue());
    await page.getByRole('button', { name: 'Play', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Pause', exact: true })).toBeVisible();
    await expect.poll(async () => Number(await range.inputValue())).toBeGreaterThan(idleTime);
    await page.getByRole('button', { name: 'Pause', exact: true }).click();

    // AC-05/AC-20: a failed switch keeps the confirmed recording and explains the failure.
    await page.route('**/api/spatial/history**', async (route) => {
      const mode = new URL(route.request().url()).searchParams.get('mode');
      if (mode === 'points') {
        await route.fulfill({ status: 500, contentType: 'application/json', body: '{}' });
        return;
      }
      await route.continue();
    });
    await page.locator('.nh-session-row[data-session-id="mock-session-afternoon"]').click();
    await expect(page.getByRole('alert').filter({ hasText: 'Spatial history could not be loaded.' })).toBeVisible();
    await expect(page.locator('.nh-coverage-segment[data-session-id="mock-session-morning"]')).toHaveAttribute('aria-pressed', 'true');
    await page.unroute('**/api/spatial/history**');

    // AC-19: a session with no recorded points is explicit and offers nothing to play.
    await page.route('**/api/spatial/history**', async (route) => {
      const mode = new URL(route.request().url()).searchParams.get('mode');
      if (mode === 'points') {
        await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' });
        return;
      }
      await route.continue();
    });
    await page.locator('.nh-session-row[data-session-id="mock-session-afternoon"]').click();
    await expect(page.getByText('No recorded positions in this session.')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Play', exact: true })).toBeDisabled();
    await page.unroute('**/api/spatial/history**');

    // AC-07/FE-04: segments are buttons, activate from the keyboard, and the scrubber steps.
    await page.locator('.nh-coverage-segment[data-session-id="mock-session-evening"]').focus();
    await page.keyboard.press('Enter');
    await expect(page.locator('.nh-coverage-segment[data-session-id="mock-session-evening"]')).toHaveAttribute('aria-pressed', 'true');
    await expect(page.locator('.nh-session-row[data-session-id="mock-session-evening"]')).toHaveAttribute('aria-pressed', 'true');
    await range.focus();
    const beforeKey = Number(await range.inputValue());
    await page.keyboard.press('ArrowRight');
    await expect.poll(async () => Number(await range.inputValue())).toBeGreaterThan(beforeKey);

    // AC-15/AC-16: recordings move into the bottom sheet without losing the selection.
    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.locator('.nh-positions-rail')).toBeHidden();
    const openRecordings = page.getByRole('button', { name: /Recordings \(3\)/ });
    await openRecordings.click();
    const sheet = page.getByRole('dialog', { name: 'Recordings' });
    await expect(sheet).toBeVisible();
    await sheet.locator('.nh-session-row[data-session-id="mock-session-afternoon"]').click();
    await expect(page.getByRole('dialog')).toHaveCount(0);
    await expect(page.locator('.nh-coverage-segment[data-session-id="mock-session-afternoon"]')).toHaveAttribute('aria-pressed', 'true');
    expect(await page.locator('html').evaluate((html) => html.scrollWidth <= html.clientWidth)).toBe(true);
    await page.setViewportSize(viewerViewport);

    // AC-18: an empty range explains itself and leaves nothing to play.
    await page.route('**/api/spatial/history**', async (route) => {
      const mode = new URL(route.request().url()).searchParams.get('mode');
      if (mode === 'sessions') {
        await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' });
        return;
      }
      await route.continue();
    });
    await page.getByRole('button', { name: 'Refresh', exact: true }).click();
    await expect(page.locator('.nh-coverage-empty')).toContainText('No recordings in this range');
    await expect(page.getByRole('button', { name: 'Play', exact: true })).toHaveCount(0);
    await page.unroute('**/api/spatial/history**');

    await page.getByRole('tab',{name:'Presence',exact:true}).click();
    await page.getByRole('button',{name:'Draw',exact:true}).click();
    await page.getByRole('button',{name:'Room',exact:true}).click();
    await page.getByRole('button',{name:'Add room outline',exact:true}).click();
    await page.getByRole('button',{name:'Add furniture',exact:true}).click();
    await page.getByLabel('Item name',{exact:true}).fill('Reading sofa');
    await page.getByLabel('Item name',{exact:true}).press('Tab');
    await page.getByRole('button',{name:'Save spatial settings',exact:true}).click();
    await expect.poll(()=>savedWorkspace?.objects[0].name).toBe('Reading sofa');
    expect(savedWorkspace?.outline).toHaveLength(4);
    await page.getByRole('button',{name:'Entry / exit',exact:true}).click();
    const entryHelp=page.locator('summary').filter({hasText:'Help with Entry / Exit'});
    await entryHelp.focus();await entryHelp.press('Enter');
    await expect(page.getByText('It is not a people counter',{exact:false})).toBeVisible();
    await page.locator('summary').filter({hasText:/^Target mapping$/}).click();
    await page.getByRole('button',{name:'Draw',exact:true}).click();
    await page.getByRole('button',{name:'Zones',exact:true}).click();
    await page.getByRole('combobox',{name:'Zone slots',exact:true}).selectOption({label:'Desk · detection'});
    await page.locator('circle.nh-handle').first().scrollIntoViewIfNeeded();
    const canvasPoint=async(x:number,y:number)=>page.locator('svg.nh-spatial-canvas').evaluate((element,point)=>{
      const position=new DOMPoint(point.x,-point.y).matrixTransform((element as SVGSVGElement).getScreenCTM()!);
      return {x:position.x,y:position.y};
    },{x,y});
    const start=await canvasPoint(-500,1000),crossed=await canvasPoint(1200,2200),finish=await canvasPoint(1400,2400);
    await page.mouse.move(start.x,start.y);await page.mouse.down();
    await page.mouse.move(crossed.x,crossed.y);await page.mouse.move(finish.x,finish.y);await page.mouse.up();
    await expect(page.getByLabel('Zone X (mm)',{exact:true})).toHaveValue('1000');
    await expect(page.getByLabel('Zone Y (mm)',{exact:true})).toHaveValue('2000');
    await expect(page.getByLabel('Zone width (mm)',{exact:true})).toHaveValue('400');
    await expect(page.getByLabel('Zone height (mm)',{exact:true})).toHaveValue('400');
    await page.getByRole('button',{name:'Undo last edit',exact:true}).click();
    for (const width of widths) {
      await page.setViewportSize({width,height:900});
      await page.getByRole('button',{name:'Draw',exact:true}).click();
      await page.getByRole('button',{name:'Zones',exact:true}).click();
      await expect(page.getByRole('heading',{name:'Zone editor',exact:true})).toBeVisible();
      await page.getByRole('combobox',{name:'Zone slots',exact:true}).selectOption({label:'Desk · detection'});
      await page.locator('svg.nh-spatial-canvas').scrollIntoViewIfNeeded();
      expect(await page.locator('html').evaluate(html=>html.scrollWidth<=html.clientWidth)).toBe(true);
      await page.screenshot({path:info.outputPath(`spatial-zone-editor-${width}.png`),fullPage:true});
    }
    // Every create action must add its own zone slot, not only the first detection button.
    await page.getByRole('button',{name:'Add exclusion zone',exact:true}).click();
    await page.getByRole('button',{name:'Add entry zone',exact:true}).click();
    const zoneSlots=page.getByRole('combobox',{name:'Zone slots',exact:true});
    await expect(zoneSlots.getByRole('option',{name:/exclusion/i})).toHaveCount(1);
    await expect(zoneSlots.getByRole('option',{name:/entry/i})).toHaveCount(1);
    sendUpdate!({ id: MOCK.livingRoomLight.id, availability: 'offline', data: {} });
    await expect(page.getByText('Device or hub connection is offline.')).toBeVisible();
    await expect(page.getByText(/Track \d+ \(/)).toHaveCount(0);
  });
});

test('ESPHome polling retains rows and does not erase a save failure', async ({ page }) => {
  let failLoad = false;
  let reads = 0;
  await page.route(`**/api${ESPHOME_DEVICES_PATH}**`, async (route) => {
    if (route.request().method() === 'PUT') {
      await route.fulfill({ status: 503, json: { data: null, error: { code: 'unavailable' } } });
      return;
    }
    reads++;
    await route.fulfill({ status: failLoad ? 503 : 200, json: { data: failLoad ? null : [{ id: 'fixture', name: 'Radar', address: '192.168.1.2', port: 6053 }], error: failLoad ? { code: 'unavailable' } : null } });
  });
  await signIn(page, '/devicedashboard');
  await page.getByRole('button', { name: 'Manage ESPHome' }).click();
  const dialog = page.getByRole('dialog', { name: 'Manage ESPHome devices' });
  await dialog.getByRole('button', { name: 'Edit Radar' }).click();
  await dialog.getByRole('button', { name: 'Save connection' }).click();
  await expect(dialog).toContainText('The device settings could not be saved.');
  const before = reads;
  await expect.poll(() => reads, { timeout: 8000 }).toBeGreaterThan(before);
  await expect(dialog).toContainText('The device settings could not be saved.');
  failLoad = true;
  await expect(dialog).toContainText('Showing the last loaded settings.', { timeout: 8000 });
  await expect(dialog.getByRole('button', { name: 'Edit Radar' })).toBeVisible();
});

test('ESPHome save completion cannot clear a reopened edit draft', async ({ page }) => {
  let release!: () => void;
  let started = false;
  const blocked = new Promise<void>((resolve) => { release = resolve; });
  await page.route(`**/api${ESPHOME_DEVICES_PATH}**`, async (route) => {
    if (route.request().method() === 'PUT') { started = true; await blocked; }
    await route.fulfill({ json: { data: [{ id: 'fixture', name: 'Radar', address: '192.168.1.2', port: 6053 }], error: null } });
  });
  try {
    await signIn(page, '/devicedashboard');
    const open = page.getByRole('button', { name: 'Manage ESPHome' });
    await open.click();
    const dialog = page.getByRole('dialog', { name: 'Manage ESPHome devices' });
    await dialog.getByRole('button', { name: 'Edit Radar' }).click();
    await dialog.getByRole('button', { name: 'Save connection' }).click();
    await expect.poll(() => started).toBe(true);
    await expect(dialog.getByRole('button', { name: 'Edit Radar' })).toBeDisabled();
    await dialog.getByRole('button', { name: 'Close', exact: true }).last().click();
    await open.click();
    await dialog.getByRole('button', { name: 'Edit Radar' }).click();
    await dialog.getByLabel('Name', { exact: true }).fill('New draft');
    const completed = page.waitForResponse((response) => response.request().method() === 'PUT');
    release();
    await completed;
    await expect(dialog.getByLabel('Name', { exact: true })).toHaveValue('New draft');
  } finally { release(); }
});

test('ESPHome About and unnamed management rows use discovered metadata', async ({ page }) => {
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const state = await response.json();
    const device = state.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
    device.provider = 'esphome';
    device.connection_type = 'esphome';
    device.properties = { model: 'Radar model', manufacturer: 'Fixture maker', firmware_version: '2026.10', project_name: 'Presence', project_version: '1.5' };
    await route.fulfill({ response, json: state });
  });
  await page.route(`**/api${ESPHOME_DEVICES_PATH}`, (route) => route.fulfill({ json: { data: [{ id: MOCK.livingRoomLight.id, address: '192.168.1.2', port: 6053 }], error: null } }));
  await signIn(page, `/devicepage/${MOCK.livingRoomLight.id}?tab=about`);
  await expect(page.getByRole('tabpanel')).toContainText('Radar model');
  await expect(page.getByRole('tabpanel')).toContainText('Fixture maker');
  await expect(page.getByRole('tabpanel')).toContainText('2026.10');
  await page.getByRole('tab', { name: 'Hub settings' }).click();
  await page.getByRole('button', { name: 'Manage ESPHome connection' }).click();
  const dialog = page.getByRole('dialog', { name: 'Manage ESPHome devices' });
  await expect(dialog.getByRole('button', { name: `Edit ${MOCK.livingRoomLight.name}` })).toBeVisible();
});
