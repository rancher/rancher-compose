# Prototyping....

This is very raw experiment. Allowing us to explore Docker Fig/Compose functionality with Rancher... 

##Supported

Bringing containers up:

 * From images
 * Ports/Expose
 * Environment vars
 * Commands
 * Privileged
 
Removing everything:

 - Remove all services/containers

## Still exploring
 
 - Links
 - Volumes
 - Volumes-from

## Building
If you run the `./scripts/build` command it will drop `./build/rancher-composer` binary. 

## Usage:

```
NAME:
   composer - Docker-compose 2 Rancher

USAGE:
   composer [global options] command [command options] [arguments...]

VERSION:
   0.1.0

AUTHOR:
  Rancher

COMMANDS:
   up       Bring all services up
   rm       Remove all containers and services
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --api-url            Specify the Rancher API Endpoint URL
   --access-key         Specify api access key
   --secret-key         Specify api secret key [$RANCHER_SECRET_KEY]
   -f "docker-compose.yml"  docker-compose yml file to use
   --help, -h           show help
   --version, -v        print the version
```
## Contact
For bugs, questions, comments, corrections, suggestions, etc., open an issue in
 [rancherio/rancher](//github.com/rancherio/rancher/issues) with a title starting with `[rancher-compose] `.

Or just [click here](//github.com/rancherio/rancher/issues/new?title=%5Brancher-compose%5D%20) to create a new issue.

# License
Copyright (c) 2014-2015 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

