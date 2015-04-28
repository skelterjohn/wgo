#wgo: managed workspaces on top of the go tool#

##Overview##

The wgo tool is a small wrapper around the go tool. It adds the concept of a workspace, in addition to that of GOPATH, and several new commands to manage that workspace.

###Goals###

- Provide a vendoring approach to dependency management for open source Go programs.
- Eventually be merged into the go tool itself (ha ha).

###How it works###

Workspaces and new subcommands.

####Workspaces####

A workspace is a directory that contains a directory ".gocfg" at its top level. Any wgo commands run with a working directory that is a subdirectory of the workspace (including the workspace itself) are said to be run from within that workspace.

####wgo foo####

When a wgo command is run from within a workspace, it runs the equivalent go command (by forwarding all arguments) with a modified environment: the GOPATH environment variable is prefixed with the workspace and any other gopaths listed in "W/.gocfg/gopaths".

So, if "W/.gocfg" exists, running wgo from within that workspace is the same as running go with `GOPATH=W:$GOPATH`. That is, the workspace will automatically be in the GOPATH, and have the highest precedence. Giving the highest GOPATH priority to "W" makes it so `go get` puts new packages in "W".

####.gocfg/vendor.json####

The ".gocfg/vendor.json" file will map import paths to repository revisions.

##New commands##

There are several new commands introduced to help with management of workspaces. If one of these commands is the first argument to wgo, it will run special logic associated with that command. Otherwise, it will forward all arguments directly to the go tool.

###wgo init###

The init command will create a ".gocfg" directory in the current directory. And "src", just to make things clear.

###wgo save###

The stash subcommand will find all revision numbers for all dependencies currently used by any package in the workspace, and write them to ".gocfg/vendor.json".

###wgo restore###

The restore subcommand will update all repositories in "W/src" to the revision numbers specified in ".gocfg/vendor.json".
