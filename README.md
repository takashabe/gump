# gump

Go Multiple-module repository対応のsemver versioningを行うためのツールです。
任意のmoduleについて、 `<module_path>/v0.0.1` のようなtagを発行します。

## Usage

例:
`my-repo/foo/bar` モジュールのバージョンアップを行う

```
$ cd /path/to/my-repo/foo/bar
$ gump --git-dir ../.. --gomod-dir .
```

あるいは

```
$ gump --git-dir . --gomod-dir foo/bar
```
