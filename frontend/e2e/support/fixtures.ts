import { test as base, expect, Page } from '@playwright/test';

export const HUB_API = 'http://localhost:4110/api';

export type Frame = { type: string; payload: any };

/** Records every WebSocket frame the page sends and receives (wire-contract assertions). */
export class WsRecorder {
  readonly sent: Frame[] = [];
  readonly received: Frame[] = [];

  constructor(page: Page) {
    page.on('websocket', (ws) => {
      ws.on('framesent', (f) => this.push(this.sent, f.payload));
      ws.on('framereceived', (f) => this.push(this.received, f.payload));
    });
  }

  private push(list: Frame[], raw: string | Buffer) {
    try {
      const obj = JSON.parse(String(raw));
      if (obj && typeof obj.type === 'string') list.push({ type: obj.type, payload: obj.payload });
    } catch {
      /* non-JSON frame */
    }
  }

  sentOf(type: string): Frame[] {
    return this.sent.filter((f) => f.type === type);
  }

  async waitForSent(type: string, match: (payload: any) => boolean = () => true, timeout = 7_000): Promise<Frame> {
    let found: Frame | undefined;
    await expect.poll(() => (found = this.sent.find((f) => f.type === type && match(f.payload))), { message: `frame ${type} sent`, timeout }).toBeTruthy();
    return found!;
  }

  async waitForReceived(type: string, match: (payload: any) => boolean = () => true, timeout = 7_000): Promise<Frame> {
    let found: Frame | undefined;
    await expect.poll(() => (found = this.received.find((f) => f.type === type && match(f.payload))), { message: `frame ${type} received`, timeout }).toBeTruthy();
    return found!;
  }
}

/**
 * Signs in through the real mock OAuth round trip (POST /auth/login -> provider URL ->
 * /auth/callback sets the session cookie -> redirect to /?auth=success), without clicking.
 */
export async function signIn(page: Page, path = '/'): Promise<void> {
  const res = await page.request.post(`${HUB_API}/auth/login`);
  const { url } = await res.json();
  await page.goto(url);
  await expect(page).not.toHaveURL(/auth=success/);
  if (path !== '/') await page.goto(path);
}

type Fixtures = { ws: WsRecorder };

export const test = base.extend<Fixtures>({
  ws: async ({ page }, use) => {
    await use(new WsRecorder(page));
  },
});

export { expect };

type Point = { x: number; y: number };

/**
 * Drags from one point to another the way the device would: touch events on touch devices
 * (phones), mouse events otherwise. Moves in small steps so drag libraries see a gesture.
 */
export async function dragBetween(page: Page, from: Point, to: Point, steps = 14): Promise<void> {
  const touch = await page.evaluate(() => navigator.maxTouchPoints > 0);
  const at = (i: number) => ({ x: from.x + ((to.x - from.x) * i) / steps, y: from.y + ((to.y - from.y) * i) / steps });
  if (!touch) {
    await page.mouse.move(from.x, from.y);
    await page.mouse.down();
    for (let i = 1; i <= steps; i++) await page.mouse.move(at(i).x, at(i).y);
    await page.mouse.up();
    return;
  }
  const cdp = await page.context().newCDPSession(page);
  const send = (type: 'touchStart' | 'touchMove' | 'touchEnd', p?: Point) =>
    cdp.send('Input.dispatchTouchEvent', { type, touchPoints: p ? [{ x: p.x, y: p.y }] : [] });
  await send('touchStart', from);
  for (let i = 1; i <= steps; i++) {
    await send('touchMove', at(i));
    await page.waitForTimeout(16);
  }
  await send('touchEnd');
  await cdp.detach();
}
