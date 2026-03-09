## Managing modules in `tcli`

This guide explains how to add a new module to the `tcli` command structure using a Swagger/OpenAPI JSON file.

### Table of contents

- [Quick start](#quick-start-junior-friendly)
- [Prerequisites](#prerequisites)
- [Command anatomy](#command-anatomy)
- [How sub-modules are created](#how-sub-modules-are-created-from-swagger-tags)
- [How API calls work](#how-api-calls-work)
- [Add a module (runtime)](#add-a-new-module-installed-runtime-config)
- [Add a module (source control)](#add-a-new-module-in-source-control-repo-owner-workflow)
- [Troubleshooting](#troubleshooting-common-issues)

---

### Who should use this guide

Use this guide if you are:

- adding a new Swagger/OpenAPI service as a `tcli` module
- trying to understand how module -> sub-module -> API command is generated
- building chained test flows using `-format` and pipes

### Quick start

If you only want the minimum steps to add a module:

1. Ensure tcli is installed (run `make install` if `~/.tcli` is missing)
2. Copy your Swagger JSON to `~/.tcli/data/<module>.json`
   > **Note:** The `data` folder must exist inside `~/.tcli`. Do not create a file named `data`.
3. Add module entry in `~/.tcli/modules.yaml`
4. Run `tcli` and confirm your module appears
5. Run `tcli <module>` and confirm sub-modules/commands appear

If any of these fail, see the [troubleshooting](#troubleshooting-common-issues) section at the end.

### Prerequisites

Install/build `tcli` first.

Check if config exists:

```console
ls ~/.tcli
```

If not found, run from the repository root:

```console
make install
```

This creates `$HOME/.tcli` and copies:

- `config.yaml`
- `modules.yaml`
- `data/`

Quick verify:

```console
tcli
```

You should see a list of supported modules.


### Command anatomy

Given a command like:

```console
tcli pce pce-customer createCustomer
```

The parts are:

- `tcli` -> CLI command
- `pce` -> module
- `pce-customer` -> sub-module (tag/group)
- `createCustomer` -> API operation/command

Pipelines can be chained with `|` where each command reads stdin from the previous command output.

### How sub-modules are created (from Swagger `tags`)

In `tcli`, sub-modules come from Swagger/OpenAPI operation tags.

- Module is from `modules.yaml` (`name` field)
- Sub-module is from each operation's `tags` value
- API command is from operation `operationId`

Example command:

```console
tcli pce pce-customer createCustomer
```

Mapping:

- `pce` -> module entry in `modules.yaml`
- `pce-customer` -> tag on the `createCustomer` API operation
- `createCustomer` -> operationId in Swagger JSON

Example Swagger pattern:

```json
{
  "tags": [
    { "name": "pce-customer", "description": "API for Customer Management" },
    { "name": "pce-cron", "description": "Cron job APIs" }
  ],
  "paths": {
    "/customer": {
      "post": {
        "tags": ["pce-customer"],
        "operationId": "createCustomer"
      }
    }
  }
}
```

#### Important notes

- If root-level `tags` are missing, `tcli` discovers tags from path operations.
- If an operation has multiple tags, that operation can appear under each tagged sub-module.
- If an operation has no tag, it is exposed directly as a module-level command.

#### Quick checks

```console
tcli pce
tcli pce pce-customer
```

Use these to confirm sub-modules and API commands were generated as expected.

### How API calls work

`tcli` reads Swagger/OpenAPI JSON for each module and uniquely identifies each REST API by `operationId`.

Generic pattern:

```console
tcli <module> <sub-module> <operationId>
```

#### Request mapping

Input and request behavior are derived from Swagger fields:

- `parameters` -> path, query, header, and body inputs
- HTTP verb + endpoint -> from operation under `paths`
- `consumes` (Swagger 2.0) or request `content` (OpenAPI 3) -> request content type

This is why `tcli` can map CLI input and stdin JSON to the API request shape directly from spec.

#### Test Case chaining

`tcli` supports sequential test execution and response-to-request transformation through shell pipelines.

- Sequential execution using `|`
- Response to request transformation using `-format`
- Built-in jq via `gojq` for lightweight JSON shaping

Example chaining:

```console
tcli pce pce-customer getTenantByTenantId -tenantId tenant11 -format '{tenantId:.tenantId}' | \
tcli pce pce-cams-data createCamsData
```

Flow:

1. First command fetches tenant data
2. `-format` keeps/transforms required fields (for example `tenantId`)
3. Next command consumes transformed JSON as input and executes

This helps build setup -> action -> verify style test cases with minimal external tooling.


### Add a new module (installed runtime config)

Use this path when you want the module available immediately on your machine.

#### 1) Copy your Swagger JSON

Copy your service spec into:

```text
~/.tcli/data/<your-module>.json
```

Example:

```text
~/.tcli/data/pce.json
~/.tcli/data/carepack_service.json
```

#### 2) Update module registry

Edit:

```text
~/.tcli/modules.yaml
```

Add an entry under `modules:`:

```yaml
modules:
  - name: pce
    config: data/pce.json
    description: pc entitlement Service api commands
  - name: carepack
    config: data/carepack_service.json
    description: Carepack Service Swagger
```

Notes:

- `name` is the top-level CLI module (`tcli <name> ...`)
- `config` is relative to `~/.tcli`
- Keep names lowercase and shell-friendly

#### 3) Verify module is loaded

Run:

```console
tcli
```

You should see your new module in the supported module list.

Then inspect commands in your module:

```console
tcli <your-module>
```

Then inspect API commands under a sub-module:

```console
tcli <your-module> <your-sub-module>
```

### Add a new module in source control (repo owner workflow)

Use this path when you want to share the module definition with the team.

1. Add Swagger JSON to:

```text
tools/data/<your-module>.json
```

2. Add module entry in:

```text
tools/modules.yaml
```

3. Re-install to sync runtime config:

```console
make install
```

This copies updated `tools/modules.yaml` and `tools/data/*` into `~/.tcli`.

### Quick validation checklist

- `~/.tcli` exists
- Swagger file exists in `~/.tcli/data`
- `~/.tcli/modules.yaml` has the new module entry
- `tcli` lists the module
- `tcli <module>` lists module commands/sub-modules

### Troubleshooting (common issues)

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| Module not visible | YAML syntax error | Check indentation in `~/.tcli/modules.yaml` |
| Module not visible | Wrong config path | Use relative path (e.g., `data/pce.json`) |
| No sub-modules | Missing tags | Add `tags` to operations in Swagger JSON |
| API command missing | Missing operationId | Ensure `operationId` exists for endpoint |
| Source changes not reflected | Stale install | Run `make install` to sync `~/.tcli` |

---

### See also

- [README](/README.md)
- [How to build and run](/docs/build_and_run.md)
- [Example explanation](/docs/example_explanation.md)
- [jq language reference](https://github.com/jqlang/jq/wiki/jq-Language-Description)
