import { expect, test, describe } from "bun:test";
import { Evaluator } from "../src/engine";
import type { FlagConfig, UserContext } from "../src/types";

describe("Evaluator", () => {
  const evaluator = new Evaluator();

  test("should return default variation when no rules match", () => {
    const config: FlagConfig = {
      id: "flag-1",
      key: "test-flag",
      enabled: true,
      default_variation_id: "v1",
      variations: [
        { id: "v1", key: "on", value: true },
        { id: "v2", key: "off", value: false },
      ],
      rules: [],
    };
    const context: UserContext = { id: "user-1", attributes: {} };
    const result = evaluator.evaluate(config, context);
    expect(result.value).toBe(true);
    expect(result.reason).toBe("default variation");
  });

  test("should match EQUALS rule", () => {
    const config: FlagConfig = {
      id: "flag-1",
      key: "test-flag",
      enabled: true,
      default_variation_id: "v2",
      variations: [
        { id: "v1", key: "on", value: true },
        { id: "v2", key: "off", value: false },
      ],
      rules: [
        {
          id: "rule-1",
          variation_id: "v1",
          clauses: [
            { attribute: "email", operator: "EQUALS", values: ["test@example.com"] },
          ],
        },
      ],
    };
    const context: UserContext = { id: "user-1", attributes: { email: "test@example.com" } };
    const result = evaluator.evaluate(config, context);
    expect(result.value).toBe(true);
    expect(result.reason).toBe("rule match: rule-1");
  });

  test("should handle percentage rollout consistently", () => {
    const config: FlagConfig = {
      id: "flag-1",
      key: "test-flag",
      enabled: true,
      rollout_percentage: 50,
      rollout_variation_id: "v1",
      default_variation_id: "v2",
      variations: [
        { id: "v1", key: "on", value: true },
        { id: "v2", key: "off", value: false },
      ],
      rules: [],
    };

    // User 1 should be consistent
    const ctx1: UserContext = { id: "user-1", attributes: {} };
    const res1 = evaluator.evaluate(config, ctx1);
    const res1_again = evaluator.evaluate(config, ctx1);
    expect(res1.value).toBe(res1_again.value);

    // Test a few users to ensure we get different results (statistically likely for 50%)
    let onCount = 0;
    for (let i = 0; i < 100; i++) {
      const res = evaluator.evaluate(config, { id: `user-${i}`, attributes: {} });
      if (res.value === true) onCount++;
    }
    // With 100 users and 50% rollout, we expect around 50 'on' results.
    // This is just a basic sanity check for the hashing distribution.
    expect(onCount).toBeGreaterThan(30);
    expect(onCount).toBeLessThan(70);
  });
});
