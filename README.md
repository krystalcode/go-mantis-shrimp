# Mantis Shrimp
An open source and more generic alternative to the Elastic X-Pack Alerting written in Golang.

## What is this repository for?
Mantis Shrimp is a new project that aims to provide a hub for watching data and triggering actions, just like the X-Pack Alerting plugin provided by Elastic. The stated goals of the project at the moment are:
* Version 1: provide a compatible replacement for Elastic X-Pack Alerting
* Version 2: extend beyond that by making it more generic (not specific to Elastic Search) and with more features

### Why Mantis Shrimp?
[Mantis Shrimp (stomatopod)](https://en.wikipedia.org/wiki/Mantis_shrimp) is a sea animal that "has one of the most elaborate visual systems ever discovered" including the broadest color-perception of all known animals, while it can deliver extremely fast and powerful hits with its claws. Likewise, this project aims to provide broad vision by watching data from various sources, and take actions fast.
The name is probably going to change as the scope and architecture of the project becomes clearer by the time it reaches its first stable version.

## Architecture
Even though the architecture is still a subject for debate, we aim to create an extensible and scalable microservices architecture. Currently the following components exist:
* Trigger API: for triggering evaluation of a watch.
* Trigger Scheduler: plugin for triggering evaluation of a watch (via Trigger API) in regular intervals
* Action API: for triggering an action
* Action Chat Message: plugin for sending a message to a chat application e.g. Rocket Chat, Slack, HipChat etc.

## How do I get set up?
The project is very new and we will only provide instructions for setting up a development environment for contributing at this stage.
The easiest way is via Docker. A docker-compose file is provided at the root folder of the project. Clone the project in a folder in your workspace that follows the structure of Golang workspace e.g. /my-workspace/go-mantis-shrimp/src/github.com/krystalcode/go-mantis-shrimp, enter the directory and run the containerized development environment:
```
# Amend the project root folder to your liking.
git clone https://github.com/krystalcode/go-mantis-shrimp.git /my-workspace/go-mantis-shrimp/src/github.com/krystalcode/go-mantis-shrimp
cd /my-workspace/go-mantis-shrimp/src/github.com/krystalcode/go-mantis-shrimp

# Export the location where any Elastic Search data will be stored on the host.
export DOCKER_COMPOSE_VOLUMES_DIR=/my-workspace/data
docker-compose up -d
```

## Contribution guidelines
We welcome all contribution so that we can make this a successful community-driven project. Please open an issue to discuss any ideas or bugs, or open a pull request.
