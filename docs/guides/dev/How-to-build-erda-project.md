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

