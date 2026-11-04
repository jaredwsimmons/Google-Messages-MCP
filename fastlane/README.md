fastlane documentation
----

# Installation

Make sure you have the latest version of the Xcode command line tools installed:

```sh
xcode-select --install
```

For _fastlane_ installation instructions, see [Installing _fastlane_](https://docs.fastlane.tools/#installing-fastlane)

# Available Actions

## Mac

### mac release

```sh
[bundle exec] fastlane mac release
```

Build, upload, and submit to App Store

### mac metadata

```sh
[bundle exec] fastlane mac metadata
```

Upload metadata only (no build)

### mac submit

```sh
[bundle exec] fastlane mac submit
```

Submit the current build for review

### mac privacy

```sh
[bundle exec] fastlane mac privacy
```

Set app privacy to 'not collecting data'

----

This README.md is auto-generated and will be re-generated every time [_fastlane_](https://fastlane.tools) is run.

More information about _fastlane_ can be found on [fastlane.tools](https://fastlane.tools).

The documentation of _fastlane_ can be found on [docs.fastlane.tools](https://docs.fastlane.tools).
