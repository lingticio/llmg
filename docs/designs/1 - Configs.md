# Configs

In order to make configurations reloadable, part of the configs multi-sourcable, and part of the configs environment-specific, we need to have a clear structure for our configs.

## Data Sources

There are multiple ways of reading configs for Gateway:

- Redis
- Plain configuration files
- Postgres, MySQL, etc.

To support diverged data sources, configurations must divide into two parts:

- **Static Configs**: These are the configurations that are read from the configuration files. They are static and do not change during the runtime.
- **Dynamic Configs**: These are the configurations that are read from the Redis. They are dynamic and can be changed during the runtime.

For the static configs,

- Redis Connection Strings
- Redis Keys
- DB Connection Strings (Postgres, MySQL)

## Hierarchy, Groups, Items
