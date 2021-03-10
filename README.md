# erda

# Develop
## build example
Here are some build commands.

build only:
```sh
make build MODULE_PATH=example
```

run after build:
```sh
make exec MODULE_PATH=example
```

build and run:
```sh
make run MODULE_PATH=example
```

print the compiled providers and help information
```sh
make run-ps MODULE_PATH=example
```

print the dependency graph of the configured providers
```sh
make run-g MODULE_PATH=example
```

build docker image only
```sh
make build-image MODULE_PATH=example
```

build docker image, and push to registry
```sh
# set Environment Variables:
# DOCKER_REGISTRY format like "registry.example.org/username" .
# DOCKER_REGISTRY_USERNAME set username for login registry if need.
# DOCKER_REGISTRY_PASSWORD set password for login registry if need.
make build-push-image MODULE_PATH=example
```

## build erda
build only:
```sh
make build MODULE_PATH=erda
```