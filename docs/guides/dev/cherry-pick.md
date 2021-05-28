# Overview

This document explains how cherry picks are managed on release branches within all erda-projects repositories.

A common use case for cherry picks is backporting PRs from master to release branches.

## Initiate a Cherry Pick

- Open your **MERGED** PR page on browser

- Add a comment

  This example applies a master branch PR to the release/1.0:
  ```
  /cherry-pick release/1.0
  ```

  Tips:
  If you comment to an **UNMERGED** PR, erda-bot will auto append a commit to tell you: cherry-pick to an unmerged PR is
  forbidden.

## Searching for Cherry Picks

Filter by labels: `auto-cherry-pick`

Examples:

- [Erda auto-cherry-pick](https://github.com/erda-project/erda/pulls?q=is%3Apr+label%3Aauto-cherry-pick)
