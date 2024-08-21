# Contributing

## Concepts

There are **projects** that use MongoDB and could work with FerretDB.
Those projects have some kind of integration tests and/or benchmarks (we will call them **tests** for simplicity); sometimes, a project is only a set of tests.
Names uniquely identify both projects and tests within the project.

Tests are run against a **database** identified by name.
Each database has a single canonical MongoDB URI and configuration.
Different configurations should be represented as different named databases.
Databases may be run locally (for example, with Docker Compose) or remotely.

Project tests a run against a database by a **runner** that returns **test results**.
In the simplest case, the test result contains only name, status, and unparsed output; other runner types may include additional properties like numerical measurements for benchmarks.
Different **runner types** accept different **parameters**; sometimes, they may include test names, and sometimes, names are parsed from the output.

**Project configuration** is a YAML file with a name matching the project name, containing a runner type, runner parameters, and expected test results per database name.
Multiple MongoDB URIs (canonical form, URI with invalid credentials, etc) are available in runner parameters via template variables.
Project configuration can not contain database-specific tests; instead, some tests may be expected to fail for some databases.

The dance tool runs tests described by project configuration and compares actual and expected results.
The dance tool itself does not start and stop databases, but the repository provides `docker-compose.yml` and `Taskfile.yml`, which do that.

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
export FERRETDB_IMAGE=ghcr.io/ferretdb/ferretdb-dev:main

bin/task env-up
bin/task DB=ferretdb-postgresql
```

The first command starts all databases defined in `docker-compose.yml`.
The second form may be used to start a single database by name (Docker Compose services use the same names) to save resources.
All databases use different ports, so running them all should be possible.

During FerretDB development, it is recommended to run it on the host with the same listening port as the matching Docker Compose service instead of using the above commands.
This way, the fix-build-test development cycle will be faster as it does not involve Docker image building.

Alternatively, the Docker image with the name `ferretdb-local` can be built by `task docker-local` in the FerretDB repository.
In this case, `FERRETDB_IMAGE` must be unset to use that default image name.

## Running tests

```sh
bin/task dance
bin/task dance DB=ferretdb-postgresql,mongodb CONFIG=python.yml
```

The first command runs all project configurations.
The second form can be used to run a single project configuration for some databases.
Both parameters are optional.

## Conventions

We expect most or all tests to pass when run against MongoDB; a few exceptions should have comments explaining why.
Tests failing against FerretDB should have issue links in the comments.
