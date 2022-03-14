# Contributing

The dance tool and tests run on the host; macOS, Linux, and Windows are expected to work.
Databases under test (FerretDB and MongoDB) may be running on the host or inside Docker; Docker Compose configuration is provided for convenience, but not required.
In particular, the FerretDB development cycle (fix-compile-run-dance) is faster with it running on the host as it does not involve Docker images building or PostgreSQL restarts.

## Running tests

```sh
make dance DB=ferretdb TEST=ferret
```

That command will run `ferret` tests ([configuration](https://github.com/FerretDB/dance/blob/main/tests/ferret.yml), [test](https://github.com/FerretDB/dance/tree/main/tests/ferret)) against FerretDB.
`DB` environment variable should have the value `ferretdb` or `mongodb`.
It defines what tests are expected to pass and fail.
For example, see [mongo-go-driver tests configuration](https://github.com/FerretDB/dance/blob/main/tests/mongo-go-driver.yml) (fields under `results.ferretdb` and `results.mongodb`).
`TEST` environment variable should have the value `ferret`, `mongo-go-driver`, or be empty.
It defines what test configuration to run; empty value runs all configurations.

## Starting environment with Docker Compose

```sh
make env-up DB=mongodb
```

That command will start MongoDB in Docker container.
Please note that running `make dance DB=ferretdb` after that would run tests against that MongoDB, but results would be compared against results expected for FerretDB.
In short, that would be wrong.

```sh
make env-up DB=ferretdb
```

That command will start FerretDB from `ferretdb-local` Docker image.
That image can be built by `make docker-local` command in FerretDB repository checkout.
As mentioned above, this approach is not recommended.
