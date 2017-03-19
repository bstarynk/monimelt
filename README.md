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


+ [Sqlite3](http://sqlite.org/) should be available from your Linux
distribution, with its `sqlite3` command and (on Debian and Ubuntu...)
`libsqlite3-dev` development package

#### installation of external dependencies

Once and for all, you need to install the external dependencies above,
using `gb vendor restore` or else with the following shell commands
(if you need to run them, be sure to `rm -rf vendor/` first). The `gb`
tool knows then (thru our *git-versionned*
[`vendor/manifest`](vendor/manifest) file).

    # run once
    gb vendor fetch github.com/mattn/go-sqlite3
    gb vendor fetch github.com/antonholmquist/jason
    gb vendor fetch github.com/petar/GoLLRB/llrb

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

These hooks are dumping (before `git commit`) and restoring (after
`git merge` or `git pull`) the `monimelt_state.sqlite` database
from/to its (`git`-versionned) `monimelt_state.sql` textual dump.