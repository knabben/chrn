# chrn

A CHANGELOG rotator and Github release automator.

The file is based on:
http://keepachangelog.com/en/1.0.0/

NOTE: Change release notes structure is based on [istio/test-infra](https://github.com/istio/test-infra/tree/master/toolbox/release_note_collector).


## Get the project

With the GOPATH variable setted, download the project
```
go get github.com/knabben/chrn
```

## Install dependencies

Install the dependencies

```
make install-dep
```

## Build the project

Generate the binary project

```
make build
```

## Knowing the flow

1) changelog - You need to populate the CHANGELOG.md file, you can skip this step, and populate your file by hand OR you can follow the pattern and get the file done.

2) rotate - Rotate your CHANGELOG with your new release (automatic discovery), create a PR with the bump version.

3) note - Create a Github Release with the description of your last released tag.


## Configuring your PRs

You will need to set some labels on your PRs, the first one is *release-note*, this is the main label, and is used to populate the changelog (you can change it via -l).

The second label you need to add is the grouper, so you can categorize PRs by *bugfix*  and *feature* for example. The title of the PR will be used to populate the changelog.

## Run it

The steps can be skipped or followed. Suppose you have followed the keepachangelog pattern.

### CHANGELOG.md generator

After filling the PRs with correct labels, pass as arguments the name of user, the repository and the local CHANGELOG.md file.

```
$ ./chrn changelog --user knabben --repo repo-test --file ~/Project/CHANGELOG.md --token token
>>> Start fetching unreleased release note from knabben/repo-test
>>> Getting PRs for [repo:knabben/repo-test label:release-note is:merged type:pr base:master merged:...]
>>> Modifying changelog file
```

You have a local modified CHANGELOG.md with the latest merged PRs title from the last release until now:

```
## [Unreleased]
## Bug
- Some PR title

## Enhancement
- Another good PR title

## [3.0.0] - 2018-02-11
```

### Commit it and bump

So lets say the latest Github release is tagged as v3.0.0, you have already modified your master and wants to release a minor version, in this case v3.1.0:

```
$ ./chrn rotate --file ~/Projects/CHANGELOG.md --org knabben --repo repo-test --token token --bump minor
```

You can see a new PR opened with the titles update. After merge or rebase on master you are ready to:

### Create a new Github release
