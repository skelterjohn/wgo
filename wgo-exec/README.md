The wgo-exec tool executes arbitrary commands, with GOPATH set as appropriate for the current wgo workspace, including those that were build from the current wgo workspace.

In a bash shell, `wgo-exec foo bar` is equivalent to running `GOPATH=$(wgo env GOPATH) foo bar`, with the PATH modified to include the various workspace bin directories. However, sometimes it's easier to change a command than to directly change the environment for a command.
