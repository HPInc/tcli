## `tcli` pipeline (YAML) schema v1

An alternative to shell-pipe scenario tests. A pipeline file describes named
steps, how data flows between them, and pipeline-wide defaults/variables.
Runnable as `tcli pipeline run <file.yaml>` and importable as a Go library
from `pkg/pipeline`.

### Why not just bash pipes?
- Bash pipelines are linear no fan-out (reuse a step's result in two
  downstream branches) and no fan-in (converge two branches into one).
- No named steps, so logs and reruns are hard to correlate.
- No shared defaults/variables `-format` strings and endpoints get
  duplicated and misquoted.
- Not portable to Windows shells.

### File shape

```yaml
name: <string>                # pipeline name (required)
description: <string>         # human summary (optional)

variables:                    # scalar reusable values
  <key>: <value>              #   referenced as ${{ variables.<key> }}

concurrency:                  # inter-step parallelism controls (optional)
  maxParallel: <int>          # cap simultaneous steps; unlimited if unset
  onFailure: cancel|drain     # cancel (default): abort in-flight siblings on
                              #   first failure; drain: let them finish

defaults:                     # applied to every step; step-level overrides win
  verbose: <bool>
  ignoreErrors: <bool>
  retryCount: <int>
  parallelism: <int>          # see the Parallelism section below
  statusCode: <int|string>
  basePath: <string>          # endpoint overrides; usually per-module but
  scheme: <string>            #   settable pipeline-wide when a whole run
  server: <string>            #   targets a non-default host
  jwt: <string>               # bearer token; prefer ${{ env.<VAR> }}

steps:                        # ordered list; execution order is a DAG derived
                              # from inputFrom + ${{ steps.* }} references +
                              # explicit dependsOn
  - name: <string>            # unique step id (required)
    command: <string>         # "<module> [<submodule>] <operation>" (required)
    params: {...}             # tcli flags/parameters for the operation
    body: <object|string>     # sugar for params.body; object is auto-serialized
    format: <string>          # jq expression applied to this step's output
                              # (equivalent to tcli's -format flag)

    inputFrom: <stepName>     # feed <stepName>'s stdout as this step's stdin
                              # exactly like a shell pipe

    outputs:                  # capture named scalars from this step's response
      <var>: <jqExpr>         #   later steps read ${{ steps.<name>.outputs.<var> }}

    dependsOn: [<stepName>]   # explicit ordering; usually inferred from
                              # inputFrom / ${{ steps.* }} references

    condition: <jqExpr>       # skip this step if expression is false/null
                              #   applied against pipeline state (see below)
    continueOnError: <bool>   # do not abort the pipeline if this step fails

    # any tcli flag can be set as a top-level step field for convenience
    count: <int>
    parallelism: <int>        # unset or 1 = sequential; N>1 = N workers;
                              #   0 = auto (GOMAXPROCS, matches CLI -parallel)
    retryCount: <int>
    ignoreErrors: <bool>
    statusCode: <int|string>
    verbose: <bool>
```

### Wiring: `inputFrom` vs `${{ steps.*.outputs.* }}`

Two orthogonal mechanisms. Pick per use case.

| Mechanism | Use when | Semantics |
|---|---|---|
| `inputFrom: <step>` (+ optional `format`) | Bulk streaming; N records flow through | Equivalent to a shell pipe `A \| B`. Runner wires `A`'s stdout to `B`'s stdin. If both steps declare `parallel: true`, it streams; otherwise `B` starts after `A` completes. |
| `${{ steps.<name>.outputs.<var> }}` inside `params:` (with an `outputs:` block on the source) | A later step needs one scalar value from an earlier response | Buffers the source step's response; extracts values with jq. Automatically creates a `dependsOn` edge. |

Both mechanisms can coexist on the same step: a step can stream data in via
`inputFrom` *and* reference `${{ steps.other.outputs.token }}` in its params.

**`outputs:` on multi-record steps.** The `outputs:` jq expression is
evaluated against each response the step produces. Recommend using
`outputs:` only on single-response steps (the common create-then-reference
pattern). If a downstream step references `${{ steps.X.outputs.var }}`
where `X` produced N responses, the value is the last one captured.
For bulk data flow between multi-record steps, use `inputFrom` instead.

### Reserved keys

The following tcli control flags MUST be set via top-level `camelCase` step
fields (or under `defaults:`), NOT inside `params:`. The loader rejects them
under `params:` to prevent two spellings for the same setting:

`format`, `count`, `parallelism`, `verbose`, `retryCount`, `ignoreErrors`,
`statusCode`, `basePath`, `scheme`, `server`, `jwt`.

Everything else under `params:` maps directly to a swagger operation
parameter, using the parameter name as declared in the spec.

### Expression syntax

`${{  }}` interpolation is supported everywhere a value is expected
(string params, format strings, condition expressions). Recognized paths:

- `${{ variables.<key> }}` from top-level `variables:`
- `${{ steps.<name>.outputs.<var> }}` from a step's `outputs:` block
- `${{ steps.<name>.status }}` HTTP status code of a completed step
- `${{ steps.<name>.ok }}` boolean; true if step succeeded
- `${{ env.<VAR> }}` from process env

### Execution model

1. Parse the file. Validate step names are unique and all references resolve.
2. Build a DAG from `inputFrom`, `${{ steps.* }}` references, and explicit
   `dependsOn`. Reject cycles.
3. Run steps in topological order. Independent branches can run concurrently,
   capped by `concurrency.maxParallel` (unlimited if unset).
4. A step fails if the underlying command returns non-zero, unless the step
   has `continueOnError: true` or `ignoreErrors: true`. On first failure,
   `concurrency.onFailure` decides whether in-flight sibling steps are
   cancelled (default) or drained.
5. Pipeline exit code is 0 iff every non-`continueOnError` step succeeded.

### Parallelism

Two orthogonal axes:

| Axis | Field | What it parallelizes |
|---|---|---|
| Inter-step | `concurrency.maxParallel` (pipeline) | Independent branches of the DAG run at the same time |
| Intra-step | `parallelism: <n>` (step) | One step processes its N input records concurrently. `parallelism: 0` matches the existing `-parallel` tcli flag (auto = GOMAXPROCS). |

The two compose freely: a step can be single-threaded internally
(`parallelism` unset or `1`) but still run at the same wall-clock time
as other independent steps.

For the rest of this section, a step is called **concurrent** when
`parallelism` is unset-or-`1` (sequential) vs. anything else (workers).

#### `inputFrom` parallelism semantics

| Producer | Consumer | Behavior |
|---|---|---|
| sequential | sequential | Producer emits all records, then consumer processes them one at a time. Deterministic order. |
| sequential | concurrent | Producer runs to completion; consumer fans records out to workers. Output order nondeterministic. |
| concurrent | sequential | Producer streams; consumer processes serially in arrival order. Matches `A -parallel \| B` today. |
| concurrent | concurrent | **Streaming**: producer and consumer run concurrently, records handed off as they're produced. Output order nondeterministic. |

#### Fan-out over a stream

Two steps can both `inputFrom: A`. The runner buffers `A`'s output and
replays it to each consumer a stream can't be teed without buffering,
so plan for the memory cost if `A` produces a large volume.

#### Fanning out over a parameter grid

There is no dedicated `matrix:` field. Express a grid as records emitted
by an upstream step; the consumer's `parallelism` handles the concurrent
invocations, one per record:

```yaml
- name: grid
  command: utils echo
  format: '[["us-east-1","us-west-2","eu-west-1"][], [10,50][]]
           | {region:.[0], batchSize:.[1]}'
  # emits 6 records: {region:"us-east-1",batchSize:10},

- name: teardown
  command: petstore pet deletePet
  inputFrom: grid
  parallelism: 6                   # -> 6 concurrent invocations
```

### Minimal example (equivalent to `examples/example.sh`)

See [`examples/pipeline/petstore_crud.yaml`](/examples/pipeline/petstore_crud.yaml).

### Open questions (v1 v2)

- **Includes / templates** pull a step group from another file so a
  create/verify/delete triple can be reused across pipelines.
- **Secrets** pipeline-level `secrets:` mapped to env for HTTP auth.

### References

- [Bash example explanation](/docs/example_explanation.md)
- [How to add a module](/docs/modules.md)
- [jq language](https://github.com/jqlang/jq/wiki/jq-Language-Description)
