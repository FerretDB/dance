# dance

[![Dance](https://github.com/FerretDB/dance/actions/workflows/dance.yml/badge.svg?branch=main)](https://github.com/FerretDB/dance/actions/workflows/dance.yml)

FerretDB integration testing tool named after [Ferret war dance](https://en.wikipedia.org/wiki/Weasel_war_dance).
It runs integration tests of various software that uses MongoDB
(such as [MongoDB Go driver](https://github.com/mongodb/mongo-go-driver))
against both MongoDB and FerretDB,
and compares results with ones expected by tests configurations.
It is expected that most or all tests pass when run against MongoDB,
so we mark a few or none tests as expected failures or skips in configuration.
More tests fail (and are marked as expected failures in tests configuration) when run against FerretDB,
but their number goes down over time.

Dance also includes one additional set of integration tests (called `diff`) that is written by FerretDB developers.
It highlights the known differences between FerretDB and MongoDB,
as [noted in documentation](https://docs.ferretdb.io/docs/diff/).

See [CONTRIBUTING.md](CONTRIBUTING.md) for running instructions.
