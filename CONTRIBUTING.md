# Contributing

The dance tool and tests run on the host; macOS, Linux, and Windows are expected to work.
Databases under test (FerretDB and MongoDB) may be running on the host or inside Docker; Docker Compose configuration is provided for convenience but not required.
In particular, the FerretDB development cycle (fix-compile-run-dance) is faster with it running on the host as it does not involve Docker image building or PostgreSQL restarts.
Running FerretDB on the host is recommended for that reason.

## Cloning repository

```sh
git clone --recursive
```

Remember to use that command to clone this repository with all submodules included.

```sh
git submodule update --init
```

If you've already cloned it without submodules you can use that command
to include required submodules.

## Running tests

```sh
bin/task dance DB=ferretdb TEST=mongo-go-driver
```

That command will run `mongo-go-driver` tests against FerretDB.
`DB` environment variable should have the value `ferretdb` or `mongodb`.
It defines what tests are expected to pass and fail.
For example, see [mongo-go-driver tests configuration](https://github.com/FerretDB/dance/blob/main/tests/mongo-go-driver.yml) (fields under `results.ferretdb` and `results.mongodb`).
`TEST` environment variable should have the value `mongo-go-driver`, or be empty.
It defines what test configuration to run; empty value runs all configurations.

## Starting environment with Docker Compose

```sh
bin/task env-up DB=mongodb
```

That command will start MongoDB in Docker container.
Please note that running `bin/task dance DB=ferretdb` after that would run tests against that MongoDB, but results would be compared against results expected for FerretDB.
In short, that would be wrong.

```sh
bin/task env-up DB=ferretdb
```

That command will start FerretDB from `ferretdb-local` Docker image.

To build a local image use the `bin/task docker-local` command in the [FerretDB](https://github.com/FerretDB/FerretDB) repository.
To use a pre-built image you must set the `FERRETDB_IMAGE` environment variable, e.g. `export FERRETDB_IMAGE=ghcr.io/ferretdb/ferretdb-dev:main`.

As mentioned above, this approach is not recommended.

### Adding tests

In order add your tests to dance CI you must use the `command` runner and add your repository as a submodule.
The `command` runner will invoke any command and CLI arguments.

For example if you wanted to add your Java application to dance, you would do the following:

1. Add the submodule to the `tests` directory `git submodule add https://github.com/my-org/my-app.git`.
2. Create a shell script in the `tests` directory called `my-app-runtime.sh` with the required logic needed to run your test.
3. Create a YAML file called `my-app.yml` in the `tests` directory and provide the `args` field with the shell script so that the runner can invoke it.
4. Start the environment and test it locally before submitting a PR to ensure that it works correctly. Refer to the above [section](https://github.com/FerretDB/dance/blob/main/CONTRIBUTING.md#starting-environment-with-docker-compose) on how to start the environment.
5. Run the test locally to verify the output `bin/task dance DB=ferretdb TEST=my-app`.
6. Submit a PR to with a title of the form "Add MyApp tests".

#### Shell script

See more examples [here](https://github.com/FerretDB/dance/tree/main/tests).

```sh
#!/bin/sh
# example bash script for my-app

set -ex

# enables Maven exceptions
export MAVEN_OPTS='-ea'

mvn compile exec:java -Dexec.mainClass=com.start.Connection \
-Dexec.args="mongodb://localhost:27017/"
```

#### YAML file

```yaml
# example YAML file for my-app
---
runner: command
# dir is where the runner is executed, setting dir is only necessary if the YAML file name differs from the repository name.
dir: java-example/java
args: [../../java-example.sh]

# we expect our test to pass so set expected_pass to 1
results:
  common:
    stats:
      expected_pass: 1

  # backend specific stats
  ferretdb:
    stats:

  mongodb:
    stats:
```
