import { expect, test, describe, spyOn, beforeAll, afterAll } from "bun:test";
import { Client } from "../src/client";
import type { FlagConfig } from "../src/types";

describe("Client", () => {
  const mockFlag: FlagConfig = {
    id: "flag-1",
    key: "test-flag",
    enabled: true,
    default_variation_id: "v1",
    variations: [
      { id: "v1", key: "on", value: true },
    ],
    rules: [],
  };

  beforeAll(() => {
    // Mock global fetch
    global.fetch = (async (input: string | Request) => {
      const url = typeof input === 'string' ? input : input.url;
      if (url && url.includes("/api/v1/flags")) {
        return {
          ok: true,
          json: async () => [mockFlag],
        } as any;
      }
      return { ok: false, status: 404 } as any;
    }) as any;
  });

  test("should fetch flags on init", async () => {
    const client = new Client({
      evaluatorUrl: "http://localhost:8082",
      streamerUrl: "", // Disable streaming
      environmentId: "env-1",
      pollInterval: 0,
    });

    await client.init();
    
    const val = client.boolVariation("test-flag", { id: "user-1", attributes: {} }, false);
    expect(val).toBe(true);
    
    client.close();
  });

  test("should handle delta updates from EventSource", async () => {
    let latestES: any = null;
    (global as any).EventSource = class {
      onmessage: any = null;
      onerror: any = null;
      constructor() { latestES = this; }
      close() {}
    };

    let fetchCount = 0;
    const originalFetch = global.fetch;
    global.fetch = (async (input: string | Request) => {
      const url = typeof input === 'string' ? input : input.url;
      if (url && url.includes("/api/v1/flags")) {
        fetchCount++;
        return {
          ok: true,
          json: async () => [mockFlag],
        } as any;
      }
      return { ok: false, status: 404 } as any;
    }) as any;

    const client = new Client({
      evaluatorUrl: "http://localhost:8082",
      streamerUrl: "http://localhost:8084",
      environmentId: "env-1",
      pollInterval: 0,
    });

    await client.init();
    
    // Poll for SSE handler to be set (avoid race condition)
    for (let i = 0; i < 10 && !(latestES && latestES.onmessage); i++) {
      await new Promise(r => setTimeout(r, 5));
    }
    
    expect(fetchCount).toBe(1);

    // Initial check
    expect(client.boolVariation("test-flag", { id: "u1", attributes: {} }, false)).toBe(true);

    // Simulate Update
    const updatedFlag = { ...mockFlag, variations: [{ id: "v1", key: "on", value: false }] };
    const payload = {
      environment_id: "env-1",
      data: updatedFlag,
    };
    
    if (latestES && latestES.onmessage) {
      latestES.onmessage({ data: JSON.stringify(payload) });
    } else {
      throw new Error("SSE handler not set");
    }

    // Verify change and NO extra fetch
    expect(client.boolVariation("test-flag", { id: "u1", attributes: {} }, true)).toBe(false);
    expect(fetchCount).toBe(1);

    client.close();
    global.fetch = originalFetch;
  });
});
