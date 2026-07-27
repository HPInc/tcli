## Overview
`tcli` is a test client to help interact with services via openapi spec files.
For general build and run instructions, please see [how to build and run](/docs/build_and_run.md)

### Features
- Data driven modules.
  - Customize test flows with context sensitive modules.
- Api specs driven by openapi.
  - Create similar config for all test interactions.
- Integrated `jq` via `gojq` to eliminate extra tools.
  - Minimal process spawns in test flows reduces test time.
- Individual parallelism control for pipeline elements.
  - Simpler and quick stress tests.
- Meaningful feature tests via shell pipelines.
  - Composeability allows seamless feature tests with multiple apis.
  - Eg: Generate test data | Create | Get | Update | Delete | Verify

### Quick How to
`tcli` is designed to do feature testing including setup, test, validation and teardown
intuitively via a concise shell pipeline.

Imagine a standard CRUD api. `tcli` tests can be done as follows (ideally)

```console
tcli testdata get -count 100 \
  | tcli api create \
  | tcli api get \
  | tcli api update -body '{..update}' \
  | tcli api delete \
  | tcli api get -status_code 404

```

#### Working example
While the above example is possible, let us look at a real world scenario. We will use
`petstore` example from https://petstore.swagger.io/

`petstore` is a sample module included as it is familiar to developers as a swagger example.
Please see [modules](/docs/modules.md) on how to manage modules.

Here is working test code that stays true to the above ideal scenario. Note how this approach
reduces need for external tools to transform output from one command to fit input for the
next command.

```console
tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' \
  | tcli petstore pet addPet -format '{petId:.id}' \
  | tcli petstore pet getPetById -format '{petId:.id,api_key:.id}' \
  | tcli petstore pet deletePet -format '{petId:.message}' \
  | tcli petstore pet getPetById -status_code 404

{"code":1,"type":"error","message":"Pet not found"}
```

Look in the [examples](/examples) folder for all examples.

For a step by step explanation of how the above test works, See [step by step explanation](/docs/example_explanation.md)
Refer [how to build and run](/docs/build_and_run.md) if you have trouble building.

#### Where to go from here
If you are still interested, please go through [docs](/docs/build_and_run.md) to find ways to
combine `tcli` for meaningful tests.
To find even more sophisticated uses of `tcli`, see [stress test with tcli](/docs/stress_test.md)

### Using tcli as a library

`tcli`'s reusable packages (spec parsing, config/module loading, and command
execution) live under `pkg/` and can be imported directly by other Go
modules — no need to copy source files into your own repository:

```go
import (
	"github.com/hpinc/tcli/pkg/cmd"
	"github.com/hpinc/tcli/pkg/config"
	"github.com/hpinc/tcli/pkg/env"
	"github.com/hpinc/tcli/pkg/parser"
)
```

Custom extension commands (like the `sqs`/`sns` example under
[`examples/_pubsub`](/examples/_pubsub)) no longer need to be copied into
`internal/cmd`. Instead, implement `cmd.Command` in your own package and
register it with `cmd.RegisterCommand` from an `init()`, then blank-import
that package from your `main`:

```go
package mypubsub

func init() {
    cmd.RegisterCommand("sqs", &SqsCommand{})
}
```

```go
// in your main package
import _ "example.com/myclient/mypubsub"
```

See [`examples/_pubsub`](/examples/_pubsub) for a full working example.

#### Areas we need help with

- testing

- `-format` does not handle input beyond simple one level input like `5`, `{"data":5}` etc.
If fully supported for multi-level json input, the `-format` feature can be much more versatile.


### References
- [How to build and run](/docs/build_and_run.md)
- [How to add a module](/docs/modules.md)
- [Explanation of test code](/docs/example_explanation.md)
- [Examples](/examples/README.md)
- Parsing openapi spec
  - [pb33f/libopenapi](https://github.com/pb33f/libopenapi)
- Formatting output
  - [jq lang](https://github.com/jqlang/jq/wiki/jq-Language-Description)
  - [gojq use as library](https://github.com/itchyny/gojq?tab=readme-ov-file#usage-as-a-library)
