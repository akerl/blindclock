blindclock
=========

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/akerl/blindclock/build.yml?branch=main)](https://github.com/akerl/blindclock/actions)
[![GitHub release](https://img.shields.io/github/release/akerl/blindclock.svg)](https://github.com/akerl/blindclock/releases)
[![License](https://img.shields.io/github/license/akerl/blindclock)](https://github.com/akerl/blindclock/blob/master/LICENSE)

Display countdown for poker blinds

## Usage

The Lambda reads from a YAML config file in S3. The file should have the following contents:

```
slacktokens:
- TOKEN_A
- TOKEN_B
statebucket: s3-bucket-name
statekey: state.yml
```

Slack tokens should be Slack App signing secrets. They're used to validate that requests came from trusted Slack Apps.

State bucket and key define where to store the state cache.

## Installation

blindclock runs as an AWS Lambda. The most effective starting point for running it is to use [this Terraform module](https://registry.terraform.io/modules/armorfret/lambda-blindclock/aws/latest).

The module creates the necessary S3 buckets. See the Usage section above for the config details.

## License

blindclock is released under the MIT License. See the bundled LICENSE file for details.
