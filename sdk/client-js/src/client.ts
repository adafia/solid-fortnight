import { EventSource } from "eventsource";
import { Evaluator } from "./engine";
import type {
  FlagConfig,
  UserContext,
  EvaluationResult,
} from "./types";

export type Config = {
  evaluatorUrl: string;
  streamerUrl: string;
  environmentId: string;
  pollInterval?: number; // In milliseconds
};

export class Client {
  private config: Config;
  private evaluator: Evaluator;
  private flags: Map<string, FlagConfig> = new Map();
  private eventSource: EventSource | null = null;
  private pollTimer: Timer | null = null;

  constructor(config: Config) {
    this.config = {
      pollInterval: 5 * 60 * 1000, // 5 minutes default
      ...config,
    };
    this.evaluator = new Evaluator();
  }

  async init(): Promise<void> {
    await this.fetchFlags();
    this.startSync();
  }

  close(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
      this.pollTimer = null;
    }
  }

  private async fetchFlags(): Promise<void> {
    const url = `${this.config.evaluatorUrl}/api/v1/flags?environment_id=${this.config.environmentId}`;
    try {
      const resp = await fetch(url);
      if (!resp.ok) {
        console.error(`Failed to fetch flags: status ${resp.status}`);
        return;
      }
      const configs: FlagConfig[] = await resp.json();
      configs.forEach((f) => this.flags.set(f.key, f));
    } catch (err) {
      console.error(`Error fetching flags: ${err}`);
    }
  }

  private startSync(): void {
    this.connectStream();
    if (this.config.pollInterval && this.config.pollInterval > 0) {
      this.pollTimer = setInterval(() => this.fetchFlags(), this.config.pollInterval);
    }
  }

  private connectStream(): void {
    const url = `${this.config.streamerUrl}/stream?environment_id=${this.config.environmentId}`;
    this.eventSource = new EventSource(url);

    this.eventSource.onmessage = (event) => {
      if (event.data === "update") {
        console.log("Received update event from streamer, fetching flags...");
        this.fetchFlags();
      }
    };

    this.eventSource.onerror = (err) => {
      console.error(`SSE connection error: ${err}. Reconnecting in 5 seconds...`);
      this.eventSource?.close();
      setTimeout(() => this.connectStream(), 5000);
    };
  }

  boolVariation(key: string, context: UserContext, defaultValue: boolean): boolean {
    const res = this.evaluate(key, context);
    if (res?.value === null || res?.value === undefined) return defaultValue;
    return Boolean(res.value);
  }

  stringVariation(key: string, context: UserContext, defaultValue: string): string {
    const res = this.evaluate(key, context);
    if (res?.value === null || res?.value === undefined) return defaultValue;
    return String(res.value);
  }

  numberVariation(key: string, context: UserContext, defaultValue: number): number {
    const res = this.evaluate(key, context);
    if (res?.value === null || res?.value === undefined) return defaultValue;
    const num = Number(res.value);
    return isNaN(num) ? defaultValue : num;
  }

  jsonVariation<T = any>(key: string, context: UserContext, defaultValue: T): T {
    const res = this.evaluate(key, context);
    if (res?.value === null || res?.value === undefined) return defaultValue;
    return res.value as T;
  }

  private evaluate(key: string, context: UserContext): EvaluationResult | null {
    const config = this.flags.get(key);
    if (!config) return null;

    try {
      return this.evaluator.evaluate(config, context);
    } catch (err) {
      console.error(`Evaluation error for flag ${key}: ${err}`);
      return null;
    }
  }
}
