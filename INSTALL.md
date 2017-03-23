# Installation instructions for the [MELT monitor](http://github.com/bstarynk/monimelt) (in Go)


Author: [Basile Starynkevitch](http://starynkevitch.net/Basile/), France.
email: [`basile@starynkevitch.net`](mailto:basile@starynkevitch.net).


Read these instructions *entirely* before starting the installation :exclamation: 

## Prerequisite

You should use some Linux distribution and be reasonably familiar with
programming and simple system administration. You probably need root
access to install the missing dependencies (notably `go`, and perhaps
`sqlite3`).

## Installing Go

We won't cover how to [install](https://golang.org/doc/install) the
[Go language](http//golang.org/doc/) implementation. You need at least
Go 1.8. Its [go](https://golang.org/cmd/go/) command should be
available, and in your `$PATH`. You should have some *initialized [Go
workspace](https://golang.org/doc/code.html#Workspaces)*, that we
assume is the default `$HOME/go/` (so you have `src/`, `bin/`, `pkg/`
directories under `$HOME/go/`).


## For `go` users

You can't simply use `go install` on this repository. You *sometimes*
need to carefully pass `-buildmode=shared -linkshared`. So if you want
to use `go get` be sure to pass `-d` (for the "download only" mode).



### plugins and shared libraries support

The
[tutorial](http://blog.ralch.com/tutorial/golang-sharing-libraries/)
from Svetlin Ralchev on *Sharing Golang packages to C and Go* is a
useful read. See also this [plugin (Go 1.8) and
packages](https://groups.google.com/forum/#!topic/golang-nuts/IKh1BqrNoxI)
thread, and that [plugin
questions](https://groups.google.com/forum/#!topic/golang-nuts/swTLZyP5QK8)
one (both threads started by me, as a Go newbie).

We are using Go [plugin](https://tip.golang.org/pkg/plugin/) facilities which appeared first in [Go 1.8](https://tip.golang.org/doc/go1.8).

You first need to **compile the Go standard library as a *shared*
library** using:

    go install -buildmode=shared -linkshared std

and you might need to be root to run that command (once and for all;
you'll probably need to re-issue that command when upgrading your `go`
compiler tool).


## Installing our dependencies


### Sqlite

You should have installed [sqlite3](http://sqlite.org/) - in
development form (e.g. run as root `apt-get install sqlite3
libsqlite3-dev` on Debian like systems) (we use 3.16.2 and/or 3.17 which
is recommended). So the [SQLite command line
shell](http://sqlite.org/cli.html) is `sqlite3` and should be in your
`$PATH`.

### external Go dependencies

We depend on several **external Go packages** (including indirect
dependencies)

+ `jason`, from [github.com/antonholmquist/jason](https://github.com/antonholmquist/jason), for JSON things.

+ `go-sqlite3`, from [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3), is an [Sqlite3](http://sqlite.org/) driver.

+ `rbt`, from
[github.com/ocdogan/rbt/](https://github.com/ocdogan/rbt/]
is a red-black tree implementation (useful for the dictionnary of
symbols).

To install these dependencies, read then run our
[./get-monimelt-dependencies.sh](./get-monimelt-dependencies.sh) script
for `bash` shell. Notice that it is passing `-buildmode=shared
-loadshared` to `go get` (and this is why we don't recommend using `go
get` *manually* to install our dependencies).


## The global state Sqlite database

Most of the persistent global state of *monimelt* is kept in the
`monimelt_global.sqlite` database. We restore it (once) from its
`monimelt_global.sql` dump file (which is versionned under *git*):

    sqlite3 monimelt_global.sqlite < monimelt_global.sql

There could also be some persistent *user state* in
`monimelt_user.sql` dump (but that is not distributed, since every
user or system would have his own one) and `monimelt_user.sqlite`
database.

#### git hooks

To facilitate management of that database, we suggest adding (once) a
pre-commit and a post-merge [git
hook](https://git-scm.com/book/it/v2/Customizing-Git-Git-Hooks).

    cd .git/hooks
    ln -sv ../../post-merge-githook.sh post-merge
    ln -sv ../../pre-commit-githook.sh pre-commit
    cd ../..

These hooks are dumping (before `git commit`) and restoring (after
`git merge` or `git pull`) the `monimelt_global.sqlite` database
from/to its (`git`-versionned) `monimelt_global.sql` textual dump (you
might improve these scripts to handle your user state).


## Building `monimelt`

By convention, our Go packages named `*mo` are our internal packages
(so they need to be built with `-linkshared -buildmode=shared`). Our
main program is in `monimelt/` sub-directory.


So you should run our [`./build-monimelt.sh`](./build-monimelt.sh)
bash script.