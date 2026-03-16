# Database Schema

```mermaid

erDiagram
    PROJECTS ||--o{ ENVIRONMENTS : contains
    PROJECTS ||--o{ FLAGS : contains
    PROJECTS ||--o{ AUDIT_LOGS : tracks

    ENVIRONMENTS ||--o{ FLAG_ENVIRONMENTS : configures
    FLAGS ||--o{ FLAG_ENVIRONMENTS : "configured in"

    FLAG_ENVIRONMENTS ||--o{ TARGETING_RULES : defines
    FLAG_ENVIRONMENTS ||--o{ FLAG_VARIATIONS : has

    TARGETING_RULES ||--o{ RULE_CONDITIONS : contains

    USERS ||--o{ AUDIT_LOGS : performs

    EVALUATION_EVENTS }o--|| FLAGS : evaluates
    EVALUATION_EVENTS }o--|| ENVIRONMENTS : "occurs in"

    PROJECTS {
        uuid id PK
        string name
        string description
        timestamp created_at
        timestamp updated_at
    }

    ENVIRONMENTS {
        uuid id PK
        uuid project_id FK
        string name "dev, staging, prod"
        string key "unique key"
        int sort_order
        timestamp created_at
    }

    FLAGS {
        uuid id PK
        uuid project_id FK
        string key "unique within project"
        string name
        text description
        string type "boolean, string, number, json"
        jsonb tags
        uuid created_by FK
        timestamp created_at
        timestamp updated_at
        boolean archived
    }

    FLAG_ENVIRONMENTS {
        uuid id PK
        uuid flag_id FK
        uuid environment_id FK
        boolean enabled
        jsonb default_value
        int version
        timestamp updated_at
        uuid updated_by FK
    }

    FLAG_VARIATIONS {
        uuid id PK
        uuid flag_environment_id FK
        string key "variation identifier"
        jsonb value
        string name
        text description
    }

    TARGETING_RULES {
        uuid id PK
        uuid flag_environment_id FK
        int priority "lower = higher priority"
        string description
        jsonb variation_id_or_rollout
        timestamp created_at
    }

    RULE_CONDITIONS {
        uuid id PK
        uuid targeting_rule_id FK
        string attribute "userId, email, plan, etc"
        string operator "eq, in, gt, contains, regex"
        jsonb values
        string condition_type "user, segment, custom"
    }

    USERS {
        uuid id PK
        string email
        string name
        string role "admin, editor, viewer"
        timestamp created_at
    }

    AUDIT_LOGS {
        uuid id PK
        uuid project_id FK
        uuid user_id FK
        string action "create, update, delete"
        string resource_type "flag, environment, rule"
        uuid resource_id
        jsonb old_value
        jsonb new_value
        timestamp created_at
    }

    EVALUATION_EVENTS {
        bigserial id PK
        uuid flag_id FK
        uuid environment_id FK
        string user_id "user identifier"
        string variation_key
        jsonb context "user attributes"
        timestamp evaluated_at
    }
```

## Real-time Infrastructure (Redis)

In addition to PostgreSQL for persistent storage, Solid Fortnight uses **Redis** as a messaging backbone for real-time flag synchronization.

### Pub/Sub Channels

| Channel | Description | Payload |
| :--- | :--- | :--- |
| `environment_updates` | Broadcasts flag changes to the Streamer service | `{"environment_id": "uuid"}` |

### Role in Architecture

1.  **Management API**: Publishes an event to `environment_updates` whenever a flag's configuration or variation is modified.
2.  **Streamer Service**: Subscribes to the `environment_updates` channel and pushes SSE updates to all clients connected to the affected environment.
