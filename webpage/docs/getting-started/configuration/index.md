---
sidebar_position: 5
---

# Configuration

Configuration options can be set using Environment Variables, as an argument to the CLI, or in a configuration file.
We use the [Viper](https://github.com/spf13/viper) library to manage configuration. Viper supports JSON, TOML, YAML, HCL, and Java properties files.
The configuration file is optional and is not required for Neko to run.
If a configuration file is present, it will be read in and merged with the default configuration values.

Highest priority is given to the Environment Variables, followed by CLI arguments, and then the configuration file.

import Configuration from './help.tsx'

<Configuration />
