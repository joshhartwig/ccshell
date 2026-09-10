In this stage, you'll extend autocompletion to handle partial completions using the longest common prefix (LCP) logic.

Completing to Longest Common Prefix
When multiple executables match the user's input, and some are prefixes of others, your shell should complete to the longest common prefix of all matches.

For example, if these executables exist in PATH:

xyz_foo
xyz_foo_bar
xyz_foo_bar_baz
When the user types xyz_and presses tab, the shell completes to xyz_foo because that is the longest prefix shared by all three executables.

After the user types the next separator character, pressing tab completes to the next common prefix of the remaining matches:
