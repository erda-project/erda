# Erda Changelog

- [v1.0.1](#v101-2021-07-08)
- [v1.0.0](#v100-2021-06-09)

# v1.0.1 [2021-07-08]

## New features

- Support individuals to create organization if he/she doesn't belong to any organizations ([#592](https://github.com/erda-project/erda/pull/592))
- Support scaling application without restarting the existing instances ([#644](https://github.com/erda-project/erda/pull/644)) ([#645](https://github.com/erda-project/erda/pull/645))

## Fixed Issues

- Fix the issue that custom stages were created with null value ([#588](https://github.com/erda-project/erda/pull/588)) ([#606](https://github.com/erda-project/erda/pull/606))
- Keep the creator, assignee, create time and man hour unchanged when issue type is switched ([#610](https://github.com/erda-project/erda/pull/610)) ([#612](https://github.com/erda-project/erda/pull/612))
- Fix the logic error of job deletion under the specified namespace ([#632](https://github.com/erda-project/erda/pull/632)) ([#636](https://github.com/erda-project/erda/pull/636))
- Add guest permissions for dashboard and ticket ([#701](https://github.com/erda-project/erda/pull/701)) ([#705](https://github.com/erda-project/erda/pull/705))
- Reset flags before loop for wait step when timed out ([#715](https://github.com/erda-project/erda/pull/715))

# v1.0.0 [2021-06-09]

Erda v1.0.0 is released!

Start your Erda journey in two ways:
- [Quick start in your local machine](https://github.com/erda-project/erda/blob/master/docs/guides/quickstart/quickstart-full.md)
- [Install with K8s](https://github.com/erda-project/erda/blob/master/docs/guides/deploy/How-to-install-Erda.md)