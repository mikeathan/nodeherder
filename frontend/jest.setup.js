// Polyfill for environment variables
//  Set environment variables for tests via process.env
process.env.VITE_WS_BASE_URL = 'ws://localhost:4001/ws';
process.env.VITE_API_BASE_URL = 'http://localhost:4001';

//  Browser Globals (using a more robust check)
if (typeof window === 'undefined') {
  // Use 'any' or 'Object.assign' to bypass strict read-only checks if necessary
  global.window = global;
  global.document = {
    documentElement: { style: {} },
    createElement: () => ({
      style: {},
      getAttribute: () => null,
    }),
  };

  // Storage mocks
  const storageMock = {
    getItem: jest.fn(() => null),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
  };
  global.localStorage = storageMock;
  global.sessionStorage = storageMock;

  global.navigator = { userAgent: 'node' };
  global.location = { origin: 'http://localhost', href: 'http://localhost/' };
}
