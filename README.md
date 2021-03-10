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

## build erda
build only:
```sh
make build MODULE_PATH=erda
```