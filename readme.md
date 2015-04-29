#wgo: managed workspaces on top of the go tool#

##Overview##

The wgo tool is a small wrapper around the go tool. It adds the concept of a workspace, in addition to that of GOPATH, and several new commands to manage that workspace.

*Do not run 'wgo restore' anywhere you care about files - this tool is in alpha and might destroy everything.*

###Goals###

- Ease unnecessary confusion around how to handle the GOPATH environment variable, especially for new gophers.
- Provide a vendoring approach to dependency management for open source Go programs.
- Eventually be merged into the go tool itself (ha ha).

####Typical use####

```
$ mkdir myproject
$ cd myproject
$ wgo init --set-primary third_party
$ wgo get github.com/someone/dep
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
third_party/src/github.com/someone/dep
$ ls -a
.gocfg src third_party
$ wgo install myproj
$ ./bin/myproj
it works!
```

####Take it for a spin####

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
third_party/src/github.com/skelterjohn/wgo-example-dep
wgo-example-w $ wgo install prog
wgo-example-w $ ./bin/prog
bar
```

###How it works###

Workspaces and new subcommands.

####Workspaces####

A workspace is a directory that contains a directory ".gocfg" at its top level. Any wgo commands run with a working directory that is a subdirectory of the workspace (including the workspace itself) are said to be run from within that workspace.

####wgo foo####

When a wgo command is run from within a workspace, it runs the equivalent go command (by forwarding all arguments) with a modified environment: the GOPATH environment variable is prefixed with the workspace and any other gopaths listed in "W/.gocfg/gopaths".

So, if "W/.gocfg" exists, running wgo from within that workspace is the same as running go with each of the directories listed in "W/.gocfg/gopaths" inserted into the beginning of GOPATH, in order.

You can modify "W/.gocfg/gopaths" at any time to change the GOPATH priority. For instance, if you put third party dependencies in "W/third_party/src", and you want calls to `go get` to put new source in there, make sure "W/third_party" is the first line in "W/.gocfg/gopaths".

####.gocfg/vendor.json####

The ".gocfg/vendor.json" file maps import paths to repository revisions. It is written and used by `vendor`, which can be installed by running `go get github.com/skelterjohn/vendor`.

##New commands##

There are several new commands introduced to help with management of workspaces. If one of these commands is the first argument to wgo, it will run special logic associated with that command. Otherwise, it will forward all arguments directly to the go tool.

###wgo init###

The init command will create a ".gocfg" directory in the current directory, and ".gocfg/gopaths" within it. And "src", just to make things clear.

Extra arguments after `wgo init` will be extra directories listed in ".gocfg/gopaths". They must be relative paths, and will be interpreted as being relative to the root of the workspace.

If you provide a flag `--set-primary=DIR`, then "DIR" will be the first directory listed in ".gocfg/gopaths". Being listed first means that it will be where `go get` puts new packages, and where `wgo save` will use as a default location for packages currently outside of "W".

###wgo save###

The save subcommand will find all revision numbers for all dependencies currently used by any package in the workspace, and write them to ".gocfg/vendor.json".

Since running `wgo save` will print out a list of paths, relative to W, where it will put repositories, it makes sense to put that output into ".gitignore", ".hgignore", or whatever. Eg, `W$ wgo save >> .gitignore` is a nice convenience to make sure the repos are not accidentally included in your workspace repository, if you choose to version it.

###wgo restore###

The restore subcommand will update all repositories in "W/src" to the revision numbers specified in ".gocfg/vendor.json".

*Be wary*: At the moment, `wgo restore` will first remove the directories that are vendored. So, if you have anything in there beyond a checkout of something available from the origin, you're going to have a bad time because it's going to be deleted.
