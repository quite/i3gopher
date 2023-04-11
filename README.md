# i3gopher

`i3gopher` is a helper rodent for i3 and Sway (the i3-compatible Wayland
compositor). It may take on various chores.

## Running

You probably want to run it from your `.i3/config`. Like:

    exec --no-startup-id i3gopher

## Using

### Focus the last focused container

The rodent subscribes to i3 events and tracks the history of focused containers
*per* *workspace*. It tries to be clever and ignores containers that disappear
(are closed, or moved to another workspace).

Running `i3gopher --focus-last` will tell an already running i3gopher to focus
the last focused container in the currently focused workspace.

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
that is matched against the window's instance name (under i3). So, if you run
`i3gopher --exclude excludei3gopher`, it will exclude an Alacritty that was
started as `alacritty --class excludei3gopher`. Under Sway, Alacritty's
`--class` sets the Wayland App ID, which the exclude argument is matched
against (supported by i3wm go bindings since [this
commit](https://github.com/i3/go-i3/commit/4e3c3810804c9b631c15644c9e885c90aa1a65d7)).

## Upgrading

Since version 1.0, POSIX/GNU style flags are used:

```diff
-     -exec string    cmd to exec
-     -focus-last     focus last
+     --exec string   cmd to exec
+ -l, --focus-last    focus last
```

## TODO

- Consider floating containers. Should they be ignored? Or kept in a separate
  history? Since i3 has `focus mode_toggle` which focused the last floating or
  tiling container.

- How deep does the stack need to be?
