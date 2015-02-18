# Prototyping....

This is very raw, and allows us to explore Docker Fig/Compose functionality with Rancher

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
 
## Usage:

```
NAME:
   composer - FIG 2 Rancher

USAGE:
   composer [global options] command [command options] [arguments...]

VERSION:
   0.0.1

AUTHOR:
  Bill Maxwell - <unknown@email>

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
