 # Distributed Arithmetic Expression Calculator

## Run project:
1) Install or update [Docker](https://docs.docker.com/engine/install/)
2) Install [Git](https://git-scm.com/downloads)
3) Get the source code
```
git clone https://github.com/ShamilGudaev/YaDistributedCalculator
```
5) Change the directory
```
cd YaDistributedCalculator
```
7) Run
```
docker-compose --project-directory ./ --file ./docker/dev/compose.dev.yml up
```
The project starts at http://localhost:5173/

## Rules for expression:
1) Supported arithmetic operations `+, -, *, /`
2) Сompound expressions using parentheses brackets.

## Features:
1) The implementation is based on the principle of REST API with data transfer between services in json format.
2) The orchestrator and agents are automatically restarted when disconnected.
3) Added the possibility of monitoring agents, taking into account the number of tasks on each.
4) The frontend is implemented on Vue.js using the Event Stream principle.
5) There is no implementation of parallel calculation of a single expression on multiple agents.

## How it works:
The key endpoints of the project are shown in the diagram

<img src="/docs/scheme.jpg" width="700" height="503" >

Contact [@tosybosy](https://t.me/tosybosy/)
