# the MELT monitor (in Go)

Hosted on github as [bstarynk/monimelt](http://github.com/bstarynk/monimelt).
GPLv3 free software.

To be completed. This is a somehow a redesign and rewrite of my
[melt-monitor-2015](http://github.com/bstarynk/melt-monitor-2015) in
Go.  So the
[melt-monitor-2015/README.md](https://github.com/bstarynk/melt-monitor-2015/blob/master/README.md)
is giving *some* motivation (but a *lot* of details and design have
changed).

Compilable on Linux/Debian/x86-64 (Sid)

## Building instructions

We use [gb](https://getgb.io/), *not* `go build`, to build our
`monimelt`. The `gb` tool is able to manage *well* and download
external dependencies (into `vendor/src`). The subcommand for managing
these dependencies is `gb vendor`.

We require [Go 1.8](https://beta.golang.org/doc/go1.8) at least (on Linux/x86-64) because we need plugins.

### external dependencies

We depend on several **external packages** (including indirect dependencies)

+ `jason`, from [github.com/antonholmquist/jason](https://github.com/antonholmquist/jason), for JSON things.

+ `gosqlite`, from [github.com/gwenn/gosqlite](https://github.com/gwenn/gosqlite), is an [Sqlite3](http://sqlite.org/) driver.

+ `yacr`, from [github.com/gwenn/yacr](https://github.com/gwenn/yacr), is yet another CSV reader, it is an indirect dependency, required from `gosqlite`.

+ [Sqlite3](http://sqlite.org/) should be available from your Linux
distribution, with its `sqlite3` command and (on Debian and Ubuntu...)
`libsqlite3-dev` development package

#### installation of external dependencies

Once and for all, you need to install the external dependencies above,
using the following shell commands (if you need to run them, be
sure to `rm -rf vendor/` first). The `gb` tool knows then (thru our
*git-versionned* [`vendor/manifest`](vendor/manifest) file).

    # run once
    gb vendor fetch github.com/gwenn/yacr
    gb vendor fetch github.com/gwenn/gosqlite
    gb vendor fetch github.com/antonholmquist/jason

Actually, you could avoid doing the above, since
[`vendor/manifest`](vendor/manifest) keeps the version, repository,
revision, branch of dependencies, and **simply run `gb vendor restore`**
(from a pristine cloned repository of this `monimelt/`)

#### building the `monimelt` binary

    gb build

will also compile the dependencies when needed and build the
`monimelt` executable. The `monimelt` binary is now available on
`bin/monimelt`, but since `$HOME/bin/` is in our `$PATH` we add a
symlink

    ln -sv $PWD/bin/monimelt $HOME/bin/

#### the global state Sqlite database

Most of the persistent state of *monimelt* is kept in the
`monimelt_state.sqlite` database. We restore it (once) from its
`monimelt_state.sql` dump file (which is versionned under *git*):

    sqlite3 monimelt_state.sqlite < monimelt_state.sql

#### git hooks

To facilitate management of that database, we suggest adding (once) a
pre-commit and a post-merge [git
hook](https://git-scm.com/book/it/v2/Customizing-Git-Git-Hooks).

    cd .git/hooks
    ln -sv ../../post-merge-githook.sh post-merge
    ln -sv ../../pre-commit-githook.sh pre-commit
    cd ../..

These hooks are dumping and restoring the `monimelt_state` database.