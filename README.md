# Erda - An enterprise-grade microservice application development platform

[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
[![codecov](https://codecov.io/gh/erda-project/erda/branch/develop/graph/badge.svg?token=ZFQ3X4257K)](https://codecov.io/gh/erda-project/erda)

![](./docs/files/logo.jpg)

## Introduction

Erda is an open-source platform created by [Terminus](https://www.terminus.io/) to ensuring the development of  microservice applications. It provides DevOps, microservice governance, and multi-cloud management capabilities. The multi-cloud architecture based on Kubernetes and application-centric DevOps and microservice governance can make the development, operation, monitoring, and problem diagnosis of complex business applications simpler and more efficient.

**Functional Architecture**

![](./docs/files/functional_architecture.jpg)

We will gradually open source the entire function according to the workload. The first to complete will be DevOps, multi-cloud management, followed by microservice governance, edge computing, etc. IT service is a function planned in the roadmap, and it has not yet started.

## Architecture

We split the codes of erda into multiple repositories according to different function. The key repositories are erda, erda-proto, erda-infra, erda-ui.

**erda** It is the main repository.

[erda-proto](https://github.com/erda-project/erda-proto) Store the communication protocol definitions between erda internal services, and the componentized protocol definitions between the web front-end and back-end services.

[erda-infra](https://github.com/erda-project/erda-infra) It is a basic repository, which stores some common and basic module codes, including the wrappers of middleware SDK, etc.

[erda-ui](https://github.com/erda-project/erda-ui) It is erda's web system and an essential component of erda. Due to the separation of front-end and back-end, it is an independent repository.

## Quick start
### To start using erda

### To start developing erda

## User Documentation
- [中文](https://dice-docs.app.terminus.io)
- English

## Contributing

This section is in progress here [Contributing to Erda](/CONTRIBUTING.md)

## Contact Us

We look forward to your connecting with us, you can ask us all questions.

- Email: erda@terminus.io

## License
Erda is under the Apache 2.0 license. See the [LICENSE](/LICENSE) file for details.
