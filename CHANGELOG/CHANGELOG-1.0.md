# Erda Changelog

- [v1.0.1](#v101-2021-07-08)
- [v1.0.0](#v100-2021-06-09)

# v1.0.1 [2021-07-08]

## New features

- Support individual developer to create organization if he/she doesn't belong to any organizations ([#592](https://github.com/erda-project/erda/pull/592))
- Support scaling your application without restarting instances which already exists ([#645](https://github.com/erda-project/erda/pull/645)) ([#644](https://github.com/erda-project/erda/pull/644))

## Fixed Issues

- Fix the issue in cmdb that custom stages created with null value ([#588](https://github.com/erda-project/erda/pull/588)) ([#606](https://github.com/erda-project/erda/pull/606))
- Fix the issue in cmdb that issue retain creator&assignee&createdAt&manHour after type changing ([#610](https://github.com/erda-project/erda/pull/610)) ([#612](https://github.com/erda-project/erda/pull/612))
- Fix the issue in scheduler that update specify namespace error when deal job #632 ([#636](https://github.com/erda-project/erda/pull/636))
- Fix the issue in ui that increase guest permissions for dashboard and ticket #701 ([#705](https://github.com/erda-project/erda/pull/705))
- Fix the issue in pipeline that task loop reset flags for wait step when timeout ([#715](https://github.com/erda-project/erda/pull/715))

# v1.0.0 [2021-06-09]

Erda v1.0.0 is released!

Start your Erda journey in two ways:
- [Quick start in your local machine](https://github.com/erda-project/erda/blob/master/docs/guides/quickstart/quickstart-full.md)
- [Install with K8s](https://github.com/erda-project/erda/blob/master/docs/guides/deploy/How-to-install-Erda.md)