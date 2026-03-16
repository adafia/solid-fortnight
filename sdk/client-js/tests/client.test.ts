import { expect, test, describe, spyOn, beforeAll, afterAll } from "bun:test";
import { Client } from "../src/client";
import type { FlagConfig } from "../src/types";

// Mock EventSource
class MockEventSource {
  onmessage: ((ev: any) => void) | null = null;
  onerror: ((err: any) => void) | null = null;
  url: string;
  constructor(url: string) { this.url = url; }
  close() {}
}

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
    global.fetch = async (url: string) => {
      if (url.includes("/api/v1/flags")) {
        return {
          ok: true,
          json: async () => [mockFlag],
        } as any;
      }
      return { ok: false, status: 404 } as any;
    };
  });

  test("should fetch flags on init", async () => {
    const client = new Client({
      evaluatorUrl: "http://localhost:8082",
      streamerUrl: "http://localhost:8084",
      environmentId: "env-1",
    });

    await client.init();
    
    const val = client.boolVariation("test-flag", { id: "user-1", attributes: {} }, false);
    expect(val).toBe(true);
    
    client.close();
  });

  test("should return default value for missing flag", async () => {
    const client = new Client({
      evaluatorUrl: "http://localhost:8082",
      streamerUrl: "http://localhost:8084",
      environmentId: "env-1",
    });

    await client.init();
    
    const val = client.boolVariation("missing-flag", { id: "user-1", attributes: {} }, false);
    expect(val).toBe(false);
    
    client.close();
  });
});
