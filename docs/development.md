# Development (WIP)

*Highly* recommend you use a [dev container](https://code.visualstudio.com/docs/remote/containers) for [vscode](https://code.visualstudio.com/), I've included the `.devcontainer` I've used to develop this app. To build **n**.eko docker container run:
```
cd .docker && ./build docker
```
the `.docker` folder also contains `./test <browser>` bash script which will launch a desktop with a browser for testing out any changes with the server.

To run the client with hot loading (for development of new client features)
```
cd ./client && npm run serve
```