# Contributing

The dance tool and tests run on the host; macOS, Linux, and Windows are expected to work.
Databases under test (FerretDB and MongoDB) may be running on the host or inside Docker; Docker Compose configuration is provided for convenience, but not required.
In particular, the FerretDB development cycle (fix-compile-run-dance) is faster with it running on the host as it does not involve Docker images building or PostgreSQL restarts.

## Cloning repository
```sh
git clone --recursive
```
Remember to use that command to clone this repository with all submodules included.

---
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
You will first need to set the `FERRETDB_IMAGE` environment variable to pull the image, e.g. `export FERRETDB_IMAGE=ghcr.io/ferretdb/ferretdb-dev:main`. 
That image can be built by `bin/task docker-local` command in FerretDB repository checkout. 
As mentioned above, this approach is not recommended.
