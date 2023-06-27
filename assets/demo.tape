# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set Theme <string>              Set the theme of the terminal (JSON)
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording
#
# Sleep:
#   Sleep <time>                    Sleep for a set amount of <time> in seconds
#
# Type:
#   Type[@<time>] "<characters>"    Type <characters> into the terminal with a
#                                   <time> delay between each character
#
# Keys:
#   Backspace[@<time>] [number]     Press the Backspace key
#   Down[@<time>] [number]          Press the Down key
#   Enter[@<time>] [number]         Press the Enter key
#   Space[@<time>] [number]         Press the Space key
#   Tab[@<time>] [number]           Press the Tab key
#   Left[@<time>] [number]          Press the Left Arrow key
#   Right[@<time>] [number]         Press the Right Arrow key
#   Up[@<time>] [number]            Press the Up Arrow key
#   Down[@<time>] [number]          Press the Down Arrow key
#   Ctrl+<key>                      Press the Control key + <key> (e.g. Ctrl+C)
#
# Display:
#   Hide                            Hide the subsequent commands from the output
#   Show                            Show the subsequent commands in the output

Output assets/demo.gif
Output assets/demo.mp4
Output assets/demo.txt
# Output assets/demo.ascii

Set FontSize 24
Set Width 1200
Set Height 800

Hide


Type@0ms "export old_dir=$(pwd) new_dir=/tmp/git-repo"
Enter

Require git
Require git-cc
Require vim

Type@0ms "set -euo pipefail"
Enter

Type@0ms "rm -rf $new_dir && mkdir -p $new_dir && cd $new_dir"
Enter

Type@0ms "cp $old_dir/commit_convention.yaml $new_dir && git init && git checkout -b main"
Enter

Type@0ms "touch ./foo && git add foo"
Enter

Type@0ms "clear"
Enter

Type@0ms "export EDITOR=vi"
Enter
Show

# start the TUI
Type "git cc"
Sleep 300ms
Enter

# navigate the type-selector
Sleep 400ms
Type@250ms "fe"
Sleep 400ms
Enter

# navigate the scope-selector
Sleep 400ms
Type "scope"
Sleep 400ms
Enter

# enter a new scope
Sleep 200ms
Escape
Type "12G" # go to line 12
Enter
Type "$" # go to the end of the line
Type "i" # enter insert mode
Right
Enter
# Backspace 2
Type "- scope: description of scope"
Escape
Type ":wq"
Enter
Sleep 500ms
Enter

# describe change
Sleep 500ms
Type "informative message about the change"
Sleep 200ms
Enter

# describe a breaking change
Sleep 500ms
Type "what this commit broke"
Sleep 500ms
Enter

# fill out remaining detail
Type "3G"
Type "i"
Enter 2
Up 2
Type "additional details"
Enter
Escape
Type ":wq"
Enter
Sleep 200ms
