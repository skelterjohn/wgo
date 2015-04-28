#wgo: managed workspaces on top of the go tool#

##Overview##

The wgo tool is a small wrapper around the go tool. It adds the concept of a workspace, in addition to that of GOPATH, and several new commands to manage that workspace.

*Do not run 'wgo restore' anywhere you care about files - this tool is in alpha and might destroy everything.*

###Goals###

- Ease unnecessary confusion around how to handle the GOPATH environment variable, especially for new gophers.
- Provide a vendoring approach to dependency management for open source Go programs.
- Eventually be merged into the go tool itself (ha ha).

###How it works###

Workspaces and new subcommands.

####Workspaces####

A workspace is a directory that contains a directory ".gocfg" at its top level. Any wgo commands run with a working directory that is a subdirectory of the workspace (including the workspace itself) are said to be run from within that workspace.

####wgo foo####

When a wgo command is run from within a workspace, it runs the equivalent go command (by forwarding all arguments) with a modified environment: the GOPATH environment variable is prefixed with the workspace and any other gopaths listed in "W/.gocfg/gopaths".

So, if "W/.gocfg" exists, running wgo from within that workspace is the same as running go with `GOPATH=W:$GOPATH`. That is, the workspace will automatically be in the GOPATH, and have the highest precedence. Giving the highest GOPATH priority to "W" makes it so `go get` puts new packages in "W".

You can modify "W/.gocfg/gopaths" at any time to change the GOPATH priority. For instance, if you put third party dependencies in "W/third_party/src", and you want calls to `go get` to put new source in there, make sure "W/third_party" is the first line in "W/.gocfg/gopaths".

####.gocfg/vendor.json####

The ".gocfg/vendor.json" file maps import paths to repository revisions. It is written and used by `vendor`, which can be installed by running `go get github.com/skelterjohn/vendor`.

##New commands##

There are several new commands introduced to help with management of workspaces. If one of these commands is the first argument to wgo, it will run special logic associated with that command. Otherwise, it will forward all arguments directly to the go tool.

###wgo init###

The init command will create a ".gocfg" directory in the current directory. And "src", just to make things clear.

###wgo save###

The save subcommand will find all revision numbers for all dependencies currently used by any package in the workspace, and write them to ".gocfg/vendor.json".

###wgo restore###

The restore subcommand will update all repositories in "W/src" to the revision numbers specified in ".gocfg/vendor.json".
