### Compatibility
Instructions and documentation below are assuming linux os. All tests below are
done on `ubuntu 22.04`.
Code for `tcli` is written in `go` and should compile and run on windows and linux.
See [How to build for windows](#how-to-build-for-windows)

### How to build for linux
```
make build

$ ls -al bin/tcli
-rwxrwxr-x 1 pgp pgp 5455872 Feb 22 20:06 bin/tcli

$ file bin/tcli
bin/tcli: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, Go BuildID=_42CAo1mLeHEpXSzoeOU/whFPal09tGHMJn8UOGRY/hYG_mp_gm2R1Beti9EV5/_rlkcVLbuBsO_N7zldJs, stripped
```

### How to build for arm
```
ARCH=arm64 make build

$ ls -al bin/tcli
-rwxrwxr-x 1 pgp pgp 5177344 Feb 22 20:12 bin/tcli

$ file bin/tcli
bin/tcli: ELF 64-bit LSB executable, ARM aarch64, version 1 (SYSV), statically linked, Go BuildID=D7nU0Ob6izoHzDCD0v23/z3OkpkhlmlMLwJOASrjD/yLac6P7pUZVjYqt0yJre/z6JeUgXpjdLpHAmj8nqQ, stripped
```

### How to build for windows
```console
$ OS=windows make build

$ ls -al bin/tcli.exe
-rwxrwxr-x 1 pgp pgp 5612032 Feb 22 20:17 bin/tcli.exe

$ file bin/tcli.exe
bin/tcli.exe: PE32+ executable (console) x86-64, for MS Windows
```

If you are building on windows, you might be missing `make`, `docker` and other tools so substitute
accordingly to run the above commands.

Once built, you can copy the resultant binary to the desktop and the `tools/config.yaml` to a folder named
`.tcli` in your home directory.

### How to install (linux only)
```console
make install
```
The following changes are made to your system:
- folder `$HOME/.tcli` is created
  - if folder already exists, relevant config file(s) are backed up in the same folder
- file `/usr/local/bin/tcli` is created or updated
  - note: requires sudo if you are non root

### How to build and run from docker
The following is only tested in linux but should work similarly for windows.
This method will depend on docker and a couple of base images available. The advantage
is that you will not need other tools to build and run locally and that you can run in
complete isolation with minimal changes to your work machine.

```console
make docker
```

This will make a docker image in your local machine. Listing the image should show local.
```console
docker images tcli
REPOSITORY       TAG       IMAGE ID       CREATED       SIZE
tcli   latest    c3993db64f87   2 hours ago   30.5MB
```

At this point, you can work in two modes.
- Work in single command mode
```console
docker run --rm tcli petstore pet getPetById -petId 1

{"id":1,"category":{"id":1,"name":"string"},"name":"doggie","photoUrls":["string"],"tags":[{"id":1,"name":"string"}],"status":"available"}
```

- Work from a shell in docker.
```console
$ docker run --rm -it --entrypoint "" tcli /bin/bash
ff3b6aeb4d79:/$ tcli petstore pet getPetById -petId 1
{"id":1,"category":{"id":1,"name":"string"},"name":"doggie","photoUrls":["string"],"tags":[{"id":1,"name":"string"}],"status":"available"}
ff3b6aeb4d79:/$ exit                                                                                                                                                                        exit
```

### Override default configuration
You can override some of the default configurations.
- Config root (default is $HOME/.tcli)
  - `TCLI_CONFIG_ROOT=tools bin/tcli` (do after a `make build`)
  - This will load config, modules and data from the `tools` folder under src root.
- Config file (default is $HOME/.tcli/config.yaml)
  - `TCLI_CONFIG_FILE=/tmp/config.yaml tcli`
  - This will load from `/tmp/config.yaml`

Note: The above config override examples are linux specific. For windows,
set the environment variable accordingly.

### General command architecture
Commands support modules at the first level or as the first argument.
Each module will support a list of commands. Navigating this structure is as shown below

#### List modules
To list supported modules, invoke `tcli` without arguments.
```console
bin/tcli
Please specify a module. Supported modules are:
- petstore    Swagger Petstore
- tcp         TCP Server check
- utils       utils
```

#### List commands in a module
To list commands in a module, use the module name as the first argument
```console
bin/tcli tcp
Please specify a command. Supported commands are:
- wait_for_server       Wait for tcp server
```

#### Global command flags
All commands support a base set of arguments that are globally applicable.
Each command can further support flags that are applicable to their functionality.

```console
tcli tcp wait_for_server -help
Usage of wait_for_server:
  -base_path string
        http base path (default "/api/v1")
  -count uint
        Number of times to repeat command (default 1)
  -doc string
        Generate docs (none, shell) (default "none")
  -ignore_errors
        Ignore errors
  -parallel
        Do runs in parallel
  -retry_count uint
        Number of retries on failure (default 10)
  -scheme string
        Scheme (default "tcp")
  -server string
        Server (default "localhost:8080")
  -status_code string
        Status code to check. -1 to ignore error (default "200")
  -v    Verbose
```

#### Get portable docs on a command
Currently supports curl commands with `-doc shell`. Since the commands are generated
from current params, it offers an opportunity to troubleshoot and create doc pages from
actual working commands.
```console
bin/tcli petstore pet getPetById -petId 1 -doc shell
2025/05/15 06:05:32 curl command:
curl -X GET    https://petstore.swagger.io/v2/pet/1
{"id":1,"category":{"id":0,"name":"doggie"},"name":"doggie","photoUrls":["string"],"tags":[{"id":20,"name":"string"},{"id":30,"name":"string"}],"status":"available"}
```
By default `-v` will use a `-doc none` which will still log an http request.
`-doc shell` will log as shown above with a well formed `curl` command. This allows
debugging and tracking a header/param option quickly without spending time reading documentation.

#### Chaining commands
Command results are in `json` format. This helps combining multiple commands with minimal external tools
to create useful workflows.

Eg: using swagger petstore example, add a pet and get the added pet by id
```console
bin/tcli petstore pet addPet -body '{"name":"a","id":1}' -format '{petId: .id}' \
  | bin/tcli petstore pet getPetById

{"id":1,"name":"a","photoUrls":[],"tags":[]}
```

#### Run a command `n` times
Using the `count` flag, a command will run repeatedly for the number specified
```console
bin/tcli petstore pet getPetById -petId 1 -count 3

{"id":1,"name":"a","photoUrls":[],"tags":[]}
{"id":1,"name":"a","photoUrls":[],"tags":[]}
{"id":1,"name":"a","photoUrls":[],"tags":[]}
```
By default, commands are run sequentially.
Adding the `parallel` flag will run commands in parallel.
This will use the number of available cpu cores to set the max parallel paths.

### Packaging
`make docker` will create a docker image.

### References
- [README.md](/README.md)
- [Explanation of test code](/docs/example_explanation.md)
- [Examples](/examples/README.md)
- Format
  - [jq lang](https://github.com/jqlang/jq/wiki/jq-Language-Description)
  - [gojq use as library](https://github.com/itchyny/gojq?tab=readme-ov-file#usage-as-a-library)
