# wgo: managed workspaces on top of the go tool

## Overview
The wgo tool is a small wrapper around the go tool. It adds the concept of a workspace, in addition to that of GOPATH, and several new commands to manage that workspace.


### How is this like...

#### godep
wgo is not at all like https://github.com/tools/godep (see "Philosophy", below). However, the `wgo save --godeps` command (see "wgo save", below) makes it easy to migrate from godep to wgo.

#### gb
wgo is very much like https://github.com/constabulary/gb except for some key details.
- gb reimplements all of the build mechanics, where wgo uses the existing go tool.
- gb works only from the root of the workspace by recognizing a "src" directory, where wgo adds an additional "W/.gocfg" directory and will search directory parents to find the workspace (like git or mercurial do with ".git" or ".hg" directories). As a result, gb can use wgo workspaces, but not the other way around without first running `wgo init` in the workspace root.
- wgo has the save/restore functionality built in, where gb can potentially include them as plugins.
- Both wgo and gb make it easy for you to create a single workspace that has everything you need for your project, and therefore make it easy to put your entire project in version control in a way that is easy for others to access.


### Goals
- Ease unnecessary confusion around how to handle the GOPATH environment variable, especially for new gophers.
- Provide a vendoring approach to dependency management for open source Go programs.
- Eventually be merged into the go tool itself (ha ha).


#### Philosophy?
`go get` is not a tool to manage dependencies, it is a tool to acquire them. So, trying to cram dependency management underneath `go get` is a fundamental mistake and will only make for awkward source repos and usage.

The approach used by wgo is to go *on top* of the go tool (and `go get`). It manages your entire workspace, and uses the go commands to operate on your workspace.

As a result, github repositories that are made to work with `go get` do not work as wgo workspaces. This incongruency is intentional: `go get` fetches a single piece of your project, while wgo manages the whole thing.

The wgo tool also uses a Go-agnostic tool, vendor (github.com/skelterjohn/vfu), to manage versions of dependencies. There is no reason to restrict vendoring goodness to Go projects.

#### Typical use
```
$ mkdir myproject
$ cd myproject
$ wgo init
$ wgo get github.com/someone/dep
$ ls -a
.gocfg src vendor
$ mkdir src/myproj
$ emacs src/myproj/main.go
... import "github.com/someone/dep"
$ wgo install myproj
$ ./bin/myproj
it works!
$ git init
$ wgo save > .gitignore
$ git add .gocfg .gitignore src/myproj
$ git remote add origin https://foo.git
$ git push origin
```

And later...

```
$ git clone https://foo.git
$ cd foo
$ wgo restore
vendor/src/github.com/someone/dep
$ ls -a
.gocfg src vendor
$ wgo install myproj
$ ./bin/myproj
it works!
```


#### Take it for a spin
Repeat after me:

```
$ git clone https://github.com/skelterjohn/wgo-example-w
Cloning into 'wgo-example-w'...
remote: Counting objects: 12, done.
remote: Compressing objects: 100% (7/7), done.
remote: Total 12 (delta 0), reused 9 (delta 0), pack-reused 0
Receiving objects: 100% (12/12), done.
$ cd wgo-example-w/
wgo-example-w $ wgo restore
vendor/src/github.com/skelterjohn/wgo-example-dep
wgo-example-w $ wgo install prog
wgo-example-w $ ./bin/prog
bar
```

### How it works
Workspaces and new subcommands.


#### Workspaces
A workspace is a directory that contains a directory ".gocfg" at its top level. Any wgo commands run with a working directory that is a subdirectory of the workspace (including the workspace itself) are said to be run from within that workspace.


#### wgo foo
When a wgo command is run from within a workspace, it runs the equivalent go command (by forwarding all arguments) with a modified environment: the GOPATH environment variable is prefixed with the workspace and any other gopaths listed in "W/.gocfg/gopaths".

For `wgo get`, the GOPATH used will only be taken from "W/.gocfg/gopaths". For any other go tool command, the GOPATH will also have the value taken from wgo's environment.

So, if "W/.gocfg" exists, running wgo from within that workspace is the same as running go with each of the directories listed in "W/.gocfg/gopaths" inserted into the beginning of GOPATH, in order.

You can modify "W/.gocfg/gopaths" at any time to change the GOPATH priority. For instance, if you put third party dependencies in "W/vendor/src", and you want calls to `go get` to put new source in there, make sure "W/vendor" is the first line in "W/.gocfg/gopaths" (this is the default when you run `wgo init` with no additional arguments).


#### wgo-exec
If you install "github.com/skelterjohn/wgo/wgo-exec", the wgo-exec tool can be used to run arbitrary commands with GOPATH adjusted for the workspace. In a bash shell, running `wgo-exec foo bar` is equivalent to `GOPATH=$(wgo env GOPATH) foo bar`.

The wgo-exec tool can be useful for situations where it is easier to change the command run than to change the environment for a command.


#### .gocfg/vendor.json
The ".gocfg/vendor.json" file maps import paths to repository revisions. It is written and used by the "github.com/skelterjohn/vfu/vend" package. The `vendor` tool can also make use if it, and can be installed by running `go get github.com/skelterjohn/vfu`.


## New commands
There are several new commands introduced to help with management of workspaces. If one of these commands is the first argument to wgo, it will run special logic associated with that command. Otherwise, it will forward all arguments directly to the go tool.


### wgo init
The init command will create a ".gocfg" directory in the current directory, and ".gocfg/gopaths" within it. And "src", just to make things clear.

Extra arguments after `wgo init` will be extra directories listed in ".gocfg/gopaths". They must be relative paths, and will be interpreted as being relative to the root of the workspace.

If you provide a flag `--vendor-gopath=DIR`, then "DIR" will be the first directory listed in ".gocfg/gopaths". Being listed first means that it will be where `go get` puts new packages, and where `wgo save` will use as a default location for packages currently outside of "W".


### wgo save
The save subcommand will find all revision numbers for all dependencies currently used by any package in the workspace, and write them to ".gocfg/vendor.json".

Since running `wgo save` will print out a list of paths, relative to W, where it will put repositories, it makes sense to put that output into ".gitignore", ".hgignore", or whatever. Eg, `W$ wgo save >> .gitignore` is a nice convenience to make sure the repos are not accidentally included in your workspace repository, if you choose to version it.

Adding the `--godeps` flag after `wgo save` will cause wgo to collect revision pins from all "Godeps/Godeps.json" files it finds in the workspace, and bring them into ".gocfg/vendor.json".

As a result, a way to transform a godep-managed package into a wgo workspace is to run

```
W$ wgo init
W$ wgo get <package managed by godep>
W$ wgo save --godeps
W$ wgo restore
```


### wgo restore
The restore subcommand will update all repositories in "W/src" to the revision numbers specified in ".gocfg/vendor.json".


### wgo vendor
The vendor subcommand will find all Go dependencies that are outside of the workspace and copy them into the workspace. Useful if you intend to completely vendor a workspace.


### wgo purge
The purge subcommand lists and deletes (if you provide the `--confirm` flag) all directories that do not contain source imported by something outside of the directories being purged. By default, the first `GOPATH` is purged (and by default, that is the `vendor` dir).
