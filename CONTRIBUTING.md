# Contributing

This document describes how to run and develop dance tools and tests.

## Types of tests

You can find tests in the `tests` directory.
We support tests of different types:

* `mongo-go-driver` - to test compatibility with the MongoDB Go driver;
* `diff` - to demonstrate FerretDB-specific differences;
* `mongo-tools` - to test compatibility with the tools (like dump and restore).

## How to roll out dance locally

Clone this repository with all submodules included:

```sh
git clone --recursive
```

If you have already cloned the repo without submodules you can run this command to include required submodules:

```sh
git submodule update --init
```

## Local environment

The dance tool and tests run on the host; macOS, Linux, and Windows are expected to work.
Databases under test (FerretDB and MongoDB) may be running on the host or inside Docker.
Docker Compose configuration is provided for convenience but not required.

In particular, the FerretDB development cycle (fix-compile-run-dance) is faster with it running on the host
as it does not involve Docker image building or PostgreSQL restarts.
Running FerretDB on the host is recommended for that reason.

## Development cycle

In principle, `dance` can work with any DB instance running on the port `27017`.

### Running FerretDB

An example of a typical workflow:

* You work with a FerretDB branch on your local machine.
* For that branch, you have a FerretDB instance running on port `27017` (for example, with [FerretDB](https://github.com/FerretDB/FerretDB)'s `bin/task run` command).
* When you make a change in the local branch, you stop and start the instance again to have all the changes compiled.
* Now you can run dance tests against FerretDB: `bin/task dance DB=ferretdb` (see "How to run tests" section for additional details).

If you can't run FerretDB on the host, it's possible to start it from `ferretdb-local` Docker image.
As mentioned above, this approach is not recommended under normal circumstances.

The process is as follows:

* To build a local image use the `bin/task docker-local` command in your local
  [FerretDB](https://github.com/FerretDB/FerretDB) branch.
* Alternatively, to use a pre-built image you must set the `FERRETDB_IMAGE` environment variable,
  e.g. `export FERRETDB_IMAGE=ghcr.io/ferretdb/ferretdb-dev:main`.
* Now you can run the necessary containers with dance's `bin/task env-up DB=ferretdb` command.
* FerretDB will be available on port `27017` on the docker container.
* To run tests, use the `bin/task dance DB=ferretdb` command as usual.

### Running MongoDB

* If you want to run MongoDB tests, stop the FerretDB instance (if any) and run MongoDB instance instead.
* For example, run the `bin/task env-up DB=mongodb` command to start MongoDB container.
* Now you can run dance tests with the `bin/task dance DB=mongodb` command,
  the tests will be run against the instance available on the port `27017` (which is now MongoDB).
* Please note that the command `bin/task dance DB=ferretdb` would run tests against that MongoDB instance,
  but results would be compared with expected results for FerretDB.
In short, that would be wrong.

## How to run tests

To run the tests use the following command:

```sh
bin/task dance DB=ferretdb TEST=mongo-go-driver
```

In this example we run `mongo-go-driver` tests against FerretDB.

Possible parameters:

* The `DB` environment variable should have the value `ferretdb` or `mongodb`.
  It defines what tests are expected to pass and fail.
For example, see [mongo-go-driver tests configuration](https://github.com/FerretDB/dance/blob/main/tests/mongo-go-driver.yml)
(fields under `results.ferretdb` and `results.mongodb`).
* The `TEST` environment variable should have one of the values `mongo-go-driver`, `diff`, `mongo-tools` or be empty.
It defines what test configuration to run.
Empty value runs all configurations.

### How to configure tests

Configuration of tests is stored in the `tests/*.yml` files (one file per each type of tests).
In particular, these files contain the information about expected number of failed, skipped and passed tests
for each test configuration (database).
It also defines what is considered a failed and a passed test and which tests are skipped.

The `dance` command reads the configuration file and runs the tests.
Then it compares the actual number of passed and failed tests with the number from the configuration file.

If the numbers are different, dance prints the list of unexpected tests and exits with a non-zero code.

## How to write diff tests

Diff tests demonstrate how FerretDB differs from MongoDB.

We write diff tests when we can't write "regular" integration or compat tests because FerretDB behaves differently
from MongoDB in some cases, and those cases are visible to users.
In addition, such differences must be documented
in the [FerretDB documentation](https://docs.ferretdb.io/diff/), please take a look at it for the additional context.
There is only one diff test that is not documented there - `TestDebugError`, as it tests the command `debugError`
which is used by developers, not by users.

Diff tests are located in the `tests/diff` directory and are regular Go tests.

Currently, every diff test or subtest should have two subtests - `FerretDB` and `MongoDB`.
Those subtests usually only contain asserts based on the expected behavior of each database.

Let's take a look at the following example:

```go
func TestDollarSign(t *testing.T) {
    t.Run("Insert", func(t *testing.T) {
        _, err := db.Collection("collection").InsertOne(ctx,  bson.D{{"foo$", "bar"}})

        t.Run("FerretDB", func(t *testing.T) {
            expected := mongo.CommandError{
                Code:    2,
                Name:    "BadValue",
                Message: `invalid key: "foo$" (key must not contain '$' sign)`,
            }
            AssertEqualError(t, expected, err)
        })

        t.Run("MongoDB", func(t *testing.T) {
            require.NoError(t, err)
        })
    })

    /* Further subtests */
}
```

It works with document validation feature for the `foo$` field key
and demonstrates how FerretDB validation differs from MongoDB behavior.
In case of FerretDB a document is considered  invalid if it has a key containing the `$` sign.
In case of MongoDB such document is considered valid.
The subtest `Insert` demonstrates this difference: in case of FerretDB we expect to receive a particular error,
in case of MongoDB we expect to receive no error when we insert such document.

### Configuration for diff tests

With the current configuration, we expect that for FerretDB all the subtests that match regular expression `FerretDB$` pass
and all the subtests that don't match this regular expression fail.
So, for FerretDB the number of passed tests is equal to the number of all the `FerretDB` subtests running.
The number of failed tests is calculated hierarchically: the "parental" test
and all its subtests that don't match the regular expression `FerretDB$` are considered failed.
For MongoDB, we have a similar configuration.

In the example above (`TestDollarSign`), for FerretDB, the number of passed tests is 1 (`FerretDB` subtest).
The number of failed tests is 3 (`MongoDB` subtest, `Insert`, and finally `TestDollarSign`).
