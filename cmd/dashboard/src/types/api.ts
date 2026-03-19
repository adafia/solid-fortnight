export interface Project {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface Environment {
  id: string;
  project_id: string;
  name: string;
  key: string;
  sort_order: number;
}

export interface Clause {
  attribute: string;
  operator: 'EQUALS' | 'IN' | 'CONTAINS' | 'PERCENTAGE';
  values: string[];
}

export interface Rule {
  name: string;
  priority: number;
  clauses: Clause[];
  value: any;
}

export interface FlagConfig {
  enabled: boolean;
  default_value: any;
  rules: Rule[];
}

export interface FeatureFlag {
  id: string;
  project_id: string;
  key: string;
  name: string;
  type: 'boolean' | 'string' | 'number' | 'json';
}
