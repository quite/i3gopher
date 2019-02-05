# i3gopher

`i3gopher` is a helper rodent for i3. It may take on various chores.
You probably want to run it from your `.i3/config`. Like:

    exec --no-startup-id i3gopher

## focus the last focused container

The rodent subscribes to i3 events and keeps track of the last focused
container *on* *each* *workspace*.

In its present shape, the last focused container on a workspace is marked
(`mark --add`) with the ID of that workspace. The rodent maintains such marks
per workspace.

Running `i3gopher -last` will get hold of the ID of the currently focused
workspace, and focus the container marked with that ID. Add something like this
in your config:

    bindsym Mod1+Tab exec --no-startup-id i3gopher -last

## TODO

- Moving a container to a different workspace messes up things. A container may
  suddenly have a mark indicating that it is was last focused on a completely
  different workspace.

- No thought about floating containers.

- ...
