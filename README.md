# dance

[![Dance](https://github.com/FerretDB/dance/actions/workflows/dance.yml/badge.svg?branch=main)](https://github.com/FerretDB/dance/actions/workflows/dance.yml)

FerretDB integration testing tool named after [Ferret war dance](https://en.wikipedia.org/wiki/Weasel_war_dance).
It runs integration tests of various software that uses MongoDB
(such as [MongoDB Go driver](https://github.com/mongodb/mongo-go-driver))
against both MongoDB and FerretDB with various backends,
and compares results with ones expected by tests configurations.
It is expected that most or all tests pass when run against MongoDB,
so we mark a few or none tests as expected failures or skips in configuration.
More tests fail (and are marked as expected failures in tests configuration) when run against FerretDB,
but their number goes down over time.
All failing or skipping tests should have comments with explanation, typically a GitHub issue URL.

See [CONTRIBUTING.md](CONTRIBUTING.md) for running instructions.
