# i3gopher

`i3gopher` is a helper rodent for i3. It may take on different chores.

## focus last focused

The gopher subscribes to i3 events and keeps track of the last focused
container on each workspace.

In its present shape, the last focused container on a workspaces is marked with
the nodeid of the workspace. The gopher maintains such marks per workspace.

Running `i3gopher -last` will get hold of the id of the currently focused
workspace, and focus the container marked with that id.

## TODO

- Moving a container to a different workspace messes up things. A container may
  suddenly have a mark indicating that it is was last focused on a completely
  different workspace.

- No thought about floating containers.

- ...
