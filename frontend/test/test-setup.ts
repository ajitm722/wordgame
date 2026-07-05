import "@testing-library/jest-dom";

const { WritableStream } = require("node:stream/web");
const { MessagePort, MessageChannel } = require("node:worker_threads");

(globalThis as any).WritableStream = WritableStream;
(globalThis as any).MessagePort = MessagePort;
(globalThis as any).MessageChannel = MessageChannel;
(globalThis as any).Event = Event;
(globalThis as any).EventTarget = EventTarget;

const mockServer = require("./mock-server").default;

beforeAll(() => mockServer.listen());
afterEach(() => mockServer.resetHandlers());
afterAll(() => mockServer.close());
