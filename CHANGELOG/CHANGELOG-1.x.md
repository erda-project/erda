# Erda Changelog 1.x

<table>
<tr>
  <th title="Current">1.1<sup>Current</sup></th>
  <th title="Current">1.0</th>
</tr>
<tr>
  <td valign="top">
    <b><a href="#v110">1.1.1</a></b><br/>
    <a href="#v110">1.1.0</a><br/>
  </td>
  <td valign="top">
    <a href="#v101">1.0.1</a><br/>
    <a href="#v100">1.0.0</a><br/>
  </td>
</tr>
</table>

# v1.1.1

`2021-8-5`

### Bug Fixes

* EDAS shares the same Agent with K8s. ([#1277](https://github.com/erda-project/erda/pull/1277))
* Releases functions of cloud management platform for free users. ([#810](https://github.com/erda-project/erda-ui/pull/810))
* Fixed the issue that error occured when editing custom addon of extended service. ([#813](https://github.com/erda-project/erda-ui/pull/813))

# v1.1.0

`2021-07-28`

### New Features

- Support existing clusters importing by users ([#806](https://github.com/erda-project/erda/pull/806))
- Support subscription to issue changes, to receive notifications timely when followed issue is modified ([#451](https://github.com/erda-project/erda-ui/pull/451))
- Support asynchronous import and export of manual test cases ([#380](https://github.com/erda-project/erda-ui/pull/380))
- Support auto page refresh for automated test plan ([#446](https://github.com/erda-project/erda-ui/pull/446))
- Support bug closed time viewing and filtering ([#445](https://github.com/erda-project/erda/pull/445))
- Add project-level applications to realize rapid project migration ([#350](https://github.com/erda-project/erda-ui/pull/350))
- Support page turning in Issues > Backlog ([#395](https://github.com/erda-project/erda-ui/pull/395))
- Optimize clone address of code repository ([#155](https://github.com/erda-project/erda-ui/pull/155))
- Optimize size of sliding window and description area for issue editing ([#314](https://github.com/erda-project/erda-ui/pull/314))
- Support size adjusting for table pagination ([#1031](https://github.com/erda-project/erda/pull/1031))
- Turn enter search to delayed auto search for personal dashboard ([#324](https://github.com/erda-project/erda-ui/pull/324))
- Optimize the downloaded file name and suffix format of container log: service name_timestamp.log ([#684](https://github.com/erda-project/erda/pull/684/files))
- Upgrade logo ([#688](https://github.com/erda-project/erda-ui/pull/688))

### Bug Fixes

- Safari page crashes when access Code Repository > Commit History ([#384](https://github.com/erda-project/erda-ui/pull/384))
- The list order remains unchanged after viewing MR ([#661](https://github.com/erda-project/erda/pull/661))
- The certificate file is uploaded but its name is not displayed ([#639](https://github.com/erda-project/erda-ui/pull/639))
- Canceling the edit of merge request will clear the comparison result ([#638](https://github.com/erda-project/erda-ui/pull/638))
- Failed to download files in code repository ([#588](https://github.com/erda-project/erda-ui/pull/588))
- The pipeline node shows the previously failed error ([#422](https://github.com/erda-project/erda-ui/pull/422))
- Members will automatically log out if exit the organization ([#347](https://github.com/erda-project/erda-ui/pull/347))

### Refactor

- Split out new platform services of dop, msp, cmp, ecp and admin
  - dop ([#392](https://github.com/erda-project/erda-ui/pull/392))
  - msp ([#407](https://github.com/erda-project/erda-ui/pull/407))
  - cmp ([#416](https://github.com/erda-project/erda-ui/pull/416))
  - ecp ([#419](https://github.com/erda-project/erda-ui/pull/419))
- Remove components of qa, apim, cmdb, ops and tmc
- Add core components of cluster manager
- Add cluster-dialer instead of soldier to handle inter-cluster communication
- Add a new way to define an interface using protobuf protocol, and the msp component has been migrated

# v1.0.1

`2021-07-08`

### New features

- Support individuals to create organization if he/she doesn't belong to any organizations ([#592](https://github.com/erda-project/erda/pull/592))
- Support scaling application without restarting the existing instances ([#644](https://github.com/erda-project/erda/pull/644)) ([#645](https://github.com/erda-project/erda/pull/645))

### Bug Fixes

- Fix the issue that custom stages were created with null value ([#588](https://github.com/erda-project/erda/pull/588)) ([#606](https://github.com/erda-project/erda/pull/606))
- Keep the creator, assignee, create time and man hour unchanged when issue type is switched ([#610](https://github.com/erda-project/erda/pull/610)) ([#612](https://github.com/erda-project/erda/pull/612))
- Fix the logic error of job deletion under the specified namespace ([#632](https://github.com/erda-project/erda/pull/632)) ([#636](https://github.com/erda-project/erda/pull/636))
- Add guest permissions for dashboard and ticket ([#701](https://github.com/erda-project/erda/pull/701)) ([#705](https://github.com/erda-project/erda/pull/705))
- Reset flags before loop for wait step when timed out ([#715](https://github.com/erda-project/erda/pull/715))

# v1.0.0

`2021-06-09`

Erda v1.0.0 is released!

Start your Erda journey in two ways:

- [Quick start in your local machine](https://github.com/erda-project/erda/blob/master/docs/guides/quickstart/quickstart-full.md)
- [Install with K8s](https://github.com/erda-project/erda/blob/master/docs/guides/deploy/How-to-install-Erda.md)
