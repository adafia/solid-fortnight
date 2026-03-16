import { createHash } from "crypto";
import type {
  FlagConfig,
  UserContext,
  EvaluationResult,
  Rule,
  Clause,
} from "./types";

export class Evaluator {
  evaluate(config: FlagConfig, context: UserContext): EvaluationResult {
    if (!config.enabled) {
      return this.defaultResult(config, "flag disabled");
    }

    // 1. Targeting Rules
    for (const rule of config.rules) {
      if (this.matchesRule(rule, context)) {
        return this.getVariation(
          config,
          rule.variation_id,
          `rule match: ${rule.id}`
        );
      }
    }

    // 2. Percentage Rollout
    if (
      config.rollout_percentage !== undefined &&
      config.rollout_percentage < 100.0
    ) {
      if (
        this.isUserInRollout(
          config.id,
          context.id,
          config.rollout_percentage
        )
      ) {
        if (config.rollout_variation_id) {
          return this.getVariation(
            config,
            config.rollout_variation_id,
            "user in rollout"
          );
        }
      } else {
        return this.defaultResult(config, "user not in rollout");
      }
    }

    // 3. Default Variation
    return this.defaultResult(config, "default variation");
  }

  private matchesRule(rule: Rule, context: UserContext): boolean {
    if (!rule.clauses || rule.clauses.length === 0) {
      return false;
    }
    for (const clause of rule.clauses) {
      if (!this.matchesClause(clause, context)) {
        return false;
      }
    }
    return true;
  }

  private matchesClause(clause: Clause, context: UserContext): boolean {
    let attrValue = context.attributes[clause.attribute];
    if (attrValue === undefined) {
      if (clause.attribute === "id") {
        attrValue = context.id;
      } else {
        return false;
      }
    }

    const strVal = String(attrValue);

    switch (clause.operator) {
      case "EQUALS":
      case "IN":
        return clause.values.some((v) => strVal === v);
      case "NOT_EQUALS":
      case "NOT_IN":
        return !clause.values.some((v) => strVal === v);
      case "CONTAINS":
        return clause.values.some((v) => strVal.includes(v));
      case "NOT_CONTAINS":
        return !clause.values.some((v) => strVal.includes(v));
      case "STARTS_WITH":
        return clause.values.some((v) => strVal.startsWith(v));
      case "ENDS_WITH":
        return clause.values.some((v) => strVal.endsWith(v));
      default:
        return false;
    }
  }

  private getVariation(
    config: FlagConfig,
    variationId: string,
    reason: string
  ): EvaluationResult {
    const variation = config.variations.find((v) => v.id === variationId);
    if (variation) {
      return {
        value: variation.value,
        variation_key: variation.key,
        variation_id: variation.id,
        reason,
      };
    }
    return {
      value: null,
      reason: `variation ${variationId} not found`,
    };
  }

  private defaultResult(config: FlagConfig, reason: string): EvaluationResult {
    if (config.default_variation_id) {
      const variation = config.variations.find(
        (v) => v.id === config.default_variation_id
      );
      if (variation) {
        return {
          value: variation.value,
          variation_key: variation.key,
          variation_id: variation.id,
          reason,
        };
      }
    }
    return {
      value: null,
      reason,
    };
  }

  private isUserInRollout(
    flagId: string,
    userId: string,
    percentage: number
  ): boolean {
    if (!userId) {
      userId = Math.random().toString(36).substring(7);
    }

    const hashKey = `${flagId}:${userId}`;
    const hash = createHash("md5").update(hashKey).digest();

    // Use the first 8 bytes of the hash to get a uint64
    const hashUint = hash.readBigUInt64BE(0);

    // Map the hash to a value between 0 and 100
    const userValue = Number(hashUint % 10000n) / 100.0;

    return userValue <= percentage;
  }
}
