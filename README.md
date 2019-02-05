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

Running `i3gopher -focus-last` will get hold of the ID of the currently focused
workspace, and focus the container marked with that ID. Add something like this
in your config:

    bindsym Mod1+Tab exec --no-startup-id i3gopher --focus-last

## execute command on window event

If you are, like me, fixing to get the title of the currently focused window
into the statusbar (maybe using `xtitle`), you might have a need for triggering
`i3status` to refresh itself at suitable times. Since this rodent is already
subscribing to window events, it has the feature to execute a command upon
receiving such. You can thus start `i3gopher` like so:

    exec --no-startup-id i3gopher -exec "killall -USR1 i3status"

## TODO

- Moving a container to a different workspace messes up things. A container may
  suddenly have a mark indicating that it is was last focused on a completely
  different workspace.

- No thought about floating containers.

- ...
