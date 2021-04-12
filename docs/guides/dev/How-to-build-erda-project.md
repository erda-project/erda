# How to build the Erda project

This document helps users to compile and build the Erda project in your IDE.

## Build all modules locally

Erda has the following dependencies:
- Git (v2.8.0 or higher)
- Go (v1.15 or higher)

The following command will build all erda modules. The specific build process can refer to [ci-ct.yml](/.github/workflows/ci-it.yml) and [Makefile](/Makefile).

```
make prepare && make
```

## Build a single module locally

The following command will build a specific erda module.

```
make prepare && MODULE_PATH=<module_name> make build
```
Note: Replace <module_name> with the name of the module to be build. All included module names can be found in the  [cmd](/cmd) directory.

### examples:
- build monitor module
  ```
  make prepare && MODULE_PATH=monitor make build
  ```
- build pipeline module
  ```
  make prepare && MODULE_PATH=pipeline make build
  ```