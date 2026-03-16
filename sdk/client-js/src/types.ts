export type UserContext = {
  id: string;
  attributes: Record<string, any>;
};

export type Variation = {
  id: string;
  key: string;
  value: any;
};

export type Operator =
  | "EQUALS"
  | "NOT_EQUALS"
  | "IN"
  | "NOT_IN"
  | "CONTAINS"
  | "NOT_CONTAINS"
  | "STARTS_WITH"
  | "ENDS_WITH";

export type Clause = {
  attribute: string;
  operator: Operator;
  values: string[];
};

export type Rule = {
  id: string;
  variation_id: string;
  clauses: Clause[];
};

export type FlagConfig = {
  id: string;
  key: string;
  enabled: boolean;
  default_variation_id?: string;
  rollout_variation_id?: string;
  variations: Variation[];
  rollout_percentage?: number;
  rules: Rule[];
};

export type EvaluationResult = {
  value: any;
  variation_key?: string;
  variation_id?: string;
  reason: string;
};
