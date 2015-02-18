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

Copyright Rancher Labs