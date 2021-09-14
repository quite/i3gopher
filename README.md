# i3gopher

`i3gopher` is a helper rodent for i3 and Sway (the i3-compatible Wayland
compositor). It may take on various chores.

## Running

You probably want to run it from your `.i3/config`. Like:

    exec --no-startup-id i3gopher

## Using

### Focus the last focused container

The rodent subscribes to i3 events and tracks the history of focused containers
*on* *each* *workspace*. It tries to be clever and ignores containers that
disappear (close, or move to another workspace).

Running `i3gopher --focus-last` tell an already running i3gopher to focused the
last focused container on the currently focused workspace.

    bindsym Mod1+Tab exec --no-startup-id i3gopher --focus-last

### Execute command on window event

If you are, like me, fixing to get the title of the currently focused window
into the statusbar (maybe using `xtitle`), you might have a need for triggering
`i3status` to refresh itself at suitable times. Since this rodent is already
subscribing to window events, it has the feature to execute a command upon
receiving such. You can thus start `i3gopher` like so:

    exec --no-startup-id i3gopher --exec "killall -USR1 i3status"

### Exclude windows from history

The optional `--exclude` option can be used to make i3gopher exclude certain
windows from ever being added to the history. It takes a regular expression
that is matched against the window's instance name. So, if you run `i3gopher
--exclude excludei3gopher`, it will exclude an Alacritty that was started as
`alacritty --class excludei3gopher`. Note: I don't run Wayland yet, and I'm not
sure if this works in Sway since class/instance is an X11 thing.

## Upgrading

Since version 1.0, POSIX/GNU style flags are used:

```diff
-     -exec string    cmd to exec
-     -focus-last     focus last
+     --exec string   cmd to exec
+ -l, --focus-last    focus last
```

## TODO

- No thought about floating containers.

- ...
