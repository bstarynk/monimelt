# the MELT monitor (in Go)

Hosted on github as [bstarynk/monimelt](http://github.com/bstarynk/monimelt).
GPLv3 free software.

Author: [Basile Starynkevitch](http://starynkevitch.net/Basile/), France.
email: [`basile@starynkevitch.net`](mailto:basile@starynkevitch.net).

To be completed. This is a somehow a redesign and rewrite of my
[melt-monitor-2015](http://github.com/bstarynk/melt-monitor-2015) in
Go.  So the
[melt-monitor-2015/README.md](https://github.com/bstarynk/melt-monitor-2015/blob/master/README.md)
is giving *some* motivation (but a *lot* of details and design have
changed).

Compilable on Linux/Debian/x86-64 (Sid)

## Building instructions


Once dependencies have been installed and built, we have been able to
use the standard `go build` to compile this, e.g. with
`GOPATH=$PWD:$PWD/vendor go build -v monimelt` (producing the
`./monimelt` executable)

### external dependencies

We depend on several **external packages** (including indirect dependencies)

+ `jason`, from [github.com/antonholmquist/jason](https://github.com/antonholmquist/jason), for JSON things.

+ `go-sqlite3`, from [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3), is an [Sqlite3](http://sqlite.org/) driver.


+ `rbt`, from
[github.com/ocdogan/rbt/](https://github.com/ocdogan/rbt/]
is a red-black tree implementation (useful for the dictionnary of
symbols).

+ [Sqlite3](http://sqlite.org/) should be available from your Linux
distribution, with its `sqlite3` command and (on Debian and Ubuntu...)
`libsqlite3-dev` development package

#### installation of external dependencies

Our dependencies are described as [git
submodules](https://git-scm.com/docs/git-submodule) in the
`.gitmodules` file. They might go into `vendor/` (in a way inspired by
[gb](https://getgb.io/). For example the
[go-sqlite3](https://github.com/mattn/go-sqlite3) package could go
into `vendor/src/github.com/mattn/go-sqlite3/` etc....).

You could use `git submodule update --init` to download the external
dependencies. Actually we don't recommend doing that at first, and
leave the `go` tool to download and install them.

But we recommend having some Go workspace (e.g. with `export
GOPATH=$HOME/mygoworkspace/` ...) and run *once*
`./get-monimelt-dependencies.sh` (but have a look into that small
shell script before running it)

#### building the `monimelt` binary

The `./build-monimelt.sh` script (have a look inside it) will build
the `monimelt` executable inside the current directory. if `$HOME/bin`
is in your path, you might run once

    ln -sv $PWD/monimelt $HOME/bin/

if you don't want to add `.` to your `$PATH` (which often is a
security hole).

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

These hooks are dumping (before `git commit`) and restoring (after
`git merge` or `git pull`) the `monimelt_state.sqlite` database
from/to its (`git`-versionned) `monimelt_state.sql` textual dump.