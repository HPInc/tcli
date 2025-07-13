## Overview
It is possible to expand support to include custom commands

This example shows how to extend `tcli` to support aws `sns` and `sqs`
- [pubsub module](/examples/_pubsub/modules.yaml)
- [pubsub spec](/examples/_pubsub/pubsub.json)
- [pubsub cmd support](/examples/_pubsub/cmd)

To add pubsub as a module, do the following
- add the module definition from modules.yaml to your tools/modules.yaml
- copy `pubsub.json` to `tools/data`
- copy files under `cmd` to `internal/cmd`
- copy files under `common` to `internal/common`

### Test it out

List supported commands
```console
$ TCLI_CONFIG_ROOT=tools go run cmd/main.go pubsub
Please specify a command. Supported commands are:
- send_sns      send sns message

- send_sqs      send sqs message
```

Get help on commands
```console
$ TCLI_CONFIG_ROOT=tools go run cmd/main.go pubsub send_sns -help
Usage of send_sns:
  -arn string
        arn (default "arn:aws:sns:us-east-1:000000000000:dev")
  -base_path string
        http base path
  -count uint
        Number of times to repeat command (default 1)
  -data string
        data
  -doc string
        Generate docs (none, shell) (default "none")
  -endpoint string
        endpoint (default "http://localhost:4566")
  -format string
        json format
  -ignore_errors
        Ignore errors
  -name string
        event name
  -parallel
        Do runs in parallel
  -retry_count uint
        Number of retries on failure (default 10)
  -scheme string
        Scheme (default "http")
  -server string
        Server (default "localhost:8080")
  -status_code string
        Status code to check (default "200")
  -v    Verbose
```

At this point, setup a localstack container and configure it to do tests.
