# Contributing

## Concepts

There are *projects* that use MongoDB and could work with FerretDB.
Those projects have some kind of integration tests and/or benchmarks (we will call them *tests* for simplicity); sometimes, a project is only a set of tests.
Names uniquely identify both projects and tests within the project.

Tests are run against a *database* identified by name.
Each database has a single canonical MongoDB URI and configuration.
Different configurations should be represented as different named databases.
Databases may be run locally (for example, with Docker Compose) or remotely.

Project tests a run against a database by a *runner* that returns *test results*.
In the simplest case, the test result contains only name, status, and unparsed output; other runner types may include additional properties like numerical measurements for benchmarks.
Different *runner types* accept different *parameters*; sometimes, they may include test names, and sometimes, names are parsed from the output.

*Project configuration* is a YAML file with a name matching the project name, containing a runner type, runner parameters, and expected test results per database name.
Multiple MongoDB URIs (canonical form, URI with invalid credentials, etc) are available in runner parameters via template variables.
Project configuration can not contain database-specific tests; instead, some tests may be expected to fail for some databases.

The dance tool runs tests described by project configuration and compares actual and expected results.
The dance tool itself does not start and stop databases, but the repository provides `docker-compose.yml` and `Taskfile.yml`, which do that.

## Database names

<!-- Keep in sync with configload.go and docker-compose.yml -->

Database names follow common convention by adding the following prefixes in the specific order:

1. `-dev` suffix means ["development build"](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version#hdr-Go_build_tags) -
   the slow build with the race detector, contrasting to production builds.
   We use development builds to detect data races, and other bugs and production builds to track performance.
2. `-branch` suffix means "branch build" -
   the build of the latest commit in the `main` branch, contrasting to the build of the latest release tag.
   We use branch builds to track changes over time, and release builds to serve as a baseline.
3. `-secured` suffix means authentication is enabled.

### MongoDB

* `mongodb` - the latest release
* `mongodb-secured` - the latest release with authentication enabled

### FerretDB v1

We use only production builds for FerretDB v1.

* `ferretdb-postgresql` - the latest release with PostgreSQL backend
* `ferretdb-postgresql-secured` - the latest release with PostgreSQL backend and authentication enabled
* `ferretdb-sqlite-replset` - the latest release with SQLite backend in replica set mode
* `ferretdb-sqlite-replset-secured` - the latest release with SQLite backend in replica set mode and authentication enabled

### FerretDB v2

* `ferretdb2` - production build of the latest release
* `ferretdb2-secured` - production build of the latest release with authentication enabled
* `ferretdb2-dev` - development build of the latest release
* `ferretdb2-dev-secured` - development build of the latest release with authentication enabled
* `ferretdb2-branch` - production build of the `main` branch
* `ferretdb2-dev-branch` - development build of the `main` branch

## Cloning repository

Projects are included in the repository as git submodules.

```sh
git clone --recursive
```

The existing working copy can be updated with:

```sh
git submodule update --init
```

## Starting environment

```sh
bin/task env-up
bin/task DB=ferretdb2
```

The first command starts all databases defined in `docker-compose.yml`.
The second form may be used to start a single database by name (Docker Compose services use the same names) to save resources.
All databases use different ports, so running them all should be possible.

During FerretDB development, it is recommended to run it on the host with the same listening port as the matching Docker Compose service instead of using the above commands.
This way, the fix-build-test development cycle will be faster as it does not involve Docker image building.

Alternatively, the Docker image with the name `ferretdb-local` can be built by `task docker-local` in the FerretDB repository.
In this case, `FERRETDB_IMAGE=ferretdb-local` can be set to use that image.
Similarly, you can specify `POSTGRES_DOCUMENTDB_IMAGE` for the image build from the FerretDB/documentdb repository.

## Running tests

```sh
bin/task dance
bin/task dance DB=ferretdb2,mongodb CONFIG=python-example.yml
```

The first command runs all project configurations.
The second form can be used to run a single project configuration for some databases.
Both parameters are optional.

## Conventions

We expect most or all tests to pass when run against MongoDB; a few exceptions should have comments explaining why.
Tests failing against FerretDB should have issue links in the comments.
