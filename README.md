# ecs-gen
[![License](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](./LICENSE)
[![Build Status](http://img.shields.io/travis/codesuki/ecs-gen.svg?style=flat)](https://travis-ci.org/codesuki/ecs-gen)

Inspired by [docker-gen](https://github.com/jwilder/docker-gen) ecs-gen lets you generate config files from templates using AWS ECS cluster information. [ecs-nginx-proxy](https://github.com/codesuki/ecs-nginx-proxy) uses ecs-gen to generate nginx config files.

## Installation
### Go
`go get -u github.com/codesuki/ecs-gen`

### Docker
Use the `codesuki/ecs-gen` docker image.

## Usage
```
usage: ecs-gen --cluster=CLUSTER --template=TEMPLATE --output=OUTPUT [<flags>]

docker-gen for AWS ECS.

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
  -r, --region="ap-northeast-1"  AWS region.
  -c, --cluster=CLUSTER          ECS cluster name.
  -t, --template=TEMPLATE        Path to template file.
  -o, --output=OUTPUT            Path to output file.
      --task="ecs-nginx-proxy"   Name of ECS task containing nginx.
  -s, --signal="nginx -s reload"
                                 Command to run to signal change.
  -f, --frequency=30             Time in seconds between polling. Must be >0.
      --once                     Only execute the template once and exit.
      --version                  Show application version.
```

### Using with Docker
When using the docker image directly you can set all parameters using environment variables:
* ECS_GEN_REGION
* ECS_GEN_CLUSTER
* ECS_GEN_TEMPLATE
* ECS_GEN_OUTPUT
* ECS_GEN_SIGNAL
* ECS_GEN_FREQUENCY
* ECS_GEN_ONCE

## Example
### Fill a template once
Running the following on the commandline `ecs-gen` will query the specified cluster, execute the template and exit.
`ecs-gen --once --region=ap-northeast-1 --cluster="Cluster name" --template=template.tmpl --output=output.conf`


### Continuously update a config
To keep a config up to date try a variation of the following.
`ecs-gen --signal="nginx -s reload" --cluster=my-cluster --template=nginx.tmpl --output=/etc/nginx/conf.d/default.conf`

## Template parameters
For now the available parameters are limited to things needed to make a nginx reverse proxy. If there is demand any information available from the AWS ECS API can be exposed.

```go
type Container struct {
    Host    string // VIRTUAL_HOST environment variable
    Port    string
    Address string
}
```

## TODO
* Expose more information
* Expose VIRTUAL_HOST under environment variables instead of `Host`
