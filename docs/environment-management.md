# Environment Management

This project uses environment variables for configuration. To make local development easier, two scripts are provided to load and unload these variables from a `.env` file.

## Setup

1. **`.env`**: This file is used for your local environment. **Important: Never commit this file to source control.**

## Scripts

### 1. Load Environment Variables

**Path:** `scripts/load_env.sh`

**Usage:**
To load the variables into your current shell session, you **must** use the `source` command:

```bash
source ./scripts/load_env.sh
```

**How it works:**

1. Checks for a `.env` file in the root directory.
2. Uses `set -a` (allexport) which automatically marks every variable created or modified for export to the environment of subsequent commands.
3. Sources the `.env` file.
4. Turns off `set -a` using `set +a`.

### 2. Unset Environment Variables

**Path:** `scripts/unset_env.sh`

**Usage:**
To remove the variables from your current shell session, you **must** use the `source` command:

```bash
source ./scripts/unset_env.sh
```

**How it works:**

1. Reads the `.env` file line by line.
2. Ignores comments (lines starting with `#`) and empty lines.
3. Extracts the variable name (the part before the `=` sign).
4. Calls `unset` on each variable name to remove it from the environment.

## Verification

After loading, you can verify that the variables are correctly set by using the `echo` command:

```bash
echo $DB_USER
```

After unsetting, the same command should return an empty output.
