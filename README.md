# gump

It is a tool for performing semver versioning for Go Multiple-module repository. For any module, create a tag like `<module_path>/v0.0.1` .

## Usage

```
bump up git tag version

Usage:
  gump [flags]

Flags:
  -g, --git-dir string     repository root (the .git directory) (default ".")
  -m, --gomod-dir string   go module root (the go.mod file) (default ".")
  -h, --help               help for gump
      --major              increment major version
      --minor              increment minor version
      --patch              increment patch version (default true)
  -p, --push               push tags
```

### Example

`my-repo/foo/bar` modules bump versions patch.

```
$ cd /path/to/my-repo/foo/bar
$ gump --git-dir ../.. --gomod-dir .
```

or

```
$ gump --git-dir . --gomod-dir foo/bar
```
