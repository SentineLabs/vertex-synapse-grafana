// Jest setup file for Grafana plugin tests
import 'jest-extended';

// Mock Grafana runtime
jest.mock('@grafana/runtime', () => ({
  getBackendSrv: jest.fn(() => ({
    fetch: jest.fn(),
  })),
  getTemplateSrv: jest.fn(() => ({
    replace: jest.fn((text: string) => text),
  })),
}));

// Global test utilities
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));

// Suppress console warnings in tests
const originalWarn = console.warn;
beforeAll(() => {
  console.warn = (...args: any[]) => {
    if (
      typeof args[0] === 'string' &&
      args[0].includes('ReactDOM.render is no longer supported')
    ) {
      return;
    }
    originalWarn.call(console, ...args);
  };
});

afterAll(() => {
  console.warn = originalWarn;
});
