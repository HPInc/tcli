## Explanation of example test code
This example was used in [README.md](/README.md). Following is a step by step
explanation of the below test example.

```console
tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' \
  | tcli petstore pet addPet -format '{petId:.id}' \
  | tcli petstore pet getPetById -format '{petId:.id,api_key:.id}' \
  | tcli petstore pet deletePet -format '{petId:.message}' \
  | tcli petstore pet getPetById -status_code 404 -v

{"code":1,"type":"error","message":"Pet not found"}
```

### How does it work?
`tcli` includes built in [modules](/tools/modules.yaml) and an example [petstore config](/tools/data/petstore.json)
which connects to https://petstore.swagger.io/. Let's break up the above command to see how it works.

#### Setting up source - setup
First, the source of data is provided by a [utils config](/tools/data/utils.json) which can be used to
source input as follows. There are many ways to source input. We will just consider
one that is self contained for now.
The expected input for `addPet` api is of the form `{"id":1,"name":"1",photourls:["1"]}`.

```console
$ tcli utils echo -format 'range(1;2) | {id:.,name:.|tostring,photourls:[.|tostring]}'

{"id":1,"name":"1","photourls":["1"]}
```

##### Explanation of format
`tcli` uses `gojq` in library form from https://github.com/itchyny/gojq to bring in `jq`
language to format. This provides the needed flexibility to shape outputs to fit any input.

Please see [jq language](https://github.com/jqlang/jq/wiki/jq-Language-Description) to
expand `-format` parameter in a variety of ways to transform output or source input.


##### Creating n input lines
At this point, if we want to create 5 inputs, Here are some different ways to change the range input

```console
$ tcli utils echo -format 'range(1;5) | {id:.,name:.|tostring,photourls:[.|tostring]}'
OR
$ tcli utils echo -format 'range(5) | .+1 | {id:.,name:.|tostring,photourls:[.|tostring]}'
OR
$ tcli utils echo -data 5 -format 'range(1;.) | {id:.,name:.|tostring,photourls:[.|tostring]}'
OR
$ echo '{"data":5}' | tcli utils echo -format 'range(1;.) | {id:.,name:.|tostring,photourls:[.|tostring]}'

{"id":1,"name":"1","photourls":["1"]}
{"id":2,"name":"2","photourls":["2"]}
{"id":3,"name":"3","photourls":["3"]}
{"id":4,"name":"4","photourls":["4"]}
{"id":5,"name":"5","photourls":["5"]}
```

#### Calling create - test step
Now that we have an input that matches `addPet` api, let's start tests.
Notice the input wrapped to a `body` before feeding to `addPet`. This is because `body` is the
request body parameter for `addPet` POST api. If it appears as a json element in input, `tcli`
will fill that parameter in from input. This will allow combining tests seamlessly.

```console
tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' | \
tcli petstore pet addPet

{"id":1,"name":"1","photoUrls":[],"tags":[]}
```

#### Calling get for the created data - test continued

```console
tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' | \
tcli petstore pet addPet -format '{petId:.id}'

{"petId":1}
```
Now we are ready to chain without additional tools

```console
tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' | \
tcli petstore pet addPet -format '{petId:.id}' | \
tcli petstore pet getPetById

{"id":1,"name":"1","photoUrls":[],"tags":[]}
```

Effectively, here a pet is added, which returns `{"id":1...}`, it is transformed to `{"petId":1}`
and fed to the get api which returns the record we just added. So far so good. Let's take it to
the next step.

#### Calling delete api - test continued (cleanup)
For any good test, it is best to leave things the way we found it. A create should be paired
with a get, a get with an update, and an update with a delete. We are skipping the update for
brevity and jumping to delete.

```console
tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' | \
tcli petstore pet addPet -format '{petId:.id}' | \
tcli petstore pet getPetById -format '{petId:.id,api_key:.id}' | \
tcli petstore pet deletePet

{"code":200,"type":"unknown","message":"1"}
```

Delete is successful but notice that there is no `id` field in delete output. However, there is
a `message` field which looks like the `id`. If we play around a little bit with different `id`
values for create, we can verify that `message` field on delete is actually returning the `id`
of the deleted record. By now, `-format` is familiar and need no explanation.
Let's move to the final step and verify if the delete worked.

#### Verifying test results - final step
We have setup, test, cleanup steps done so far. Let's do verify.

```console
tcli utils echo -format 'range(1;2) | {body: {id:.,name:.|tostring,photourls:[.|tostring]}}' | \
tcli petstore pet addPet -format '{petId:.id}' | \
tcli petstore pet getPetById -format '{petId:.id,api_key:.id}' | \
tcli petstore pet deletePet -format '{petId:.message}' | \
tcli petstore pet getPetById -status_code 404

{"code":1,"type":"error","message":"Pet not found"}

$ echo $?
0
```

When `getPetById` is called after delete, `404 not found` is expected.
Since `status_code` is specified as `400`, after the tests run, exit code will be 0 for a pass.
If `status_code` is not specified in this case, it will default to `200` and tests will fail on verify.

### References
- [README.md](/README.md)
- [How to build and run](/docs/build_and_run.md)
- [Examples](/examples/README.md)
- Format
  - [jq lang](https://github.com/jqlang/jq/wiki/jq-Language-Description)
  - [gojq use as library](https://github.com/itchyny/gojq?tab=readme-ov-file#usage-as-a-library)
