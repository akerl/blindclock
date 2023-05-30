blindclock
=========

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/akerl/blindclock/build.yml?branch=main)](https://github.com/akerl/blindclock/actions)
[![GitHub release](https://img.shields.io/github/release/akerl/blindclock.svg)](https://github.com/akerl/blindclock/releases)
[![License](https://img.shields.io/github/license/akerl/blindclock)](https://github.com/akerl/blindclock/blob/master/LICENSE)

Display countdown for poker blinds

## Usage

blindclock reads from a simple config file:

```
state_file: state.yml
sqs_queue: blindclock.fifo
```

The state_file is required. This is the path where the server stores the current timer and blinds.

The sqs_queue is optional. If specified, the server will poll the SQS queue for updated state details. It sources AWS credentials using the default providers, so you'll need to set them via standard AWS SDK environment variables or configuration files.

## Installation

```
go install github.com/akerl/blindclock@latest
```

## License

blindclock is released under the MIT License. See the bundled LICENSE file for details.
