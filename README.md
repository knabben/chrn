# chrn

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

## Configuring your PRs

You will need to set some labels on your PRs, the first one is *release-note*, this is the main label, and is used to populate the changelog (you can change it via -l).

The second label you need to add is the grouper, so you can categorize PRs by *bugfix*  and *feature* for example. The title of the PR will be used to populate the changelog.

## Run it

First choose a Github user and a repository, next you will need the current_release tag (ie v4.0) and the previous_release tag (ie v3.0). This can be enough to filter data and save it on a file.

```
$ ./chrn --user knabben --repo repo-test -c v4.0 -p v3.0
2017/12/17 19:56:53 Start fetching release note from knabben/repo-test
2017/12/17 19:56:54 Query: [repo:knabben/repo-test is:merged type:pr merged:2017-12-17T13:52:38Z..2017-12-17T14:09:35Z base:master]
2017/12/17 19:56:55 Saving data on: ./release-note
```

The final format will be something like:

```
$ cat release-note
repo-test: v4.0 -- v3.0

## Release
* Adding new field - https://api.github.com/repos/knabben/repo-test/issues/8

## Bugfix
* Changing main file name - https://api.github.com/repos/knabben/repo-test/issues/9
* Correct serialization on master- https://api.github.com/repos/knabben/repo-test/issues/5
```

### Updating Github Release Note

It's possible to automatically update the notes from the release, you just need a file with your personal API token, and use it to authenticate.

```
$ ./chrn --user knabben --repo repo-test -c v4.0 -p v3.0 --token ./token --save
```
