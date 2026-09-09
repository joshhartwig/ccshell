# In this stage, you'll implement tab completion for multiple executables that share a common prefix

```Handling Multiple Matches
When a user types a command prefix and presses <TAB>, your shell should find all executables in PATH that match the prefix and do the following:

On the first <TAB> press, ring the bell (using \x07).
On the second <TAB> press:
Print all matching executables on a new line. List them in alphabetical order, separated by at least one space (two spaces are recommended for readability).
Show the prompt again on the next line, keeping the original command prefix.
For example:

$ xyz_<TAB><TAB>
xyz_bar  xyz_baz  xyz_quz
$ xyz_```

```

print -l ${^path}/*(-*N)  

```

```go
case '\t':
   // find and sort PATH exes and builtin commands
   matches := ks.completer(string(buf))
   slices.Sort(matches)

   // if we hit a second tab
   if keyState.lastKeyTab {
    if len(matches) > 1 {
     fmt.Fprintf(ks.out, "\r\n%s\r\n", strings.Join(matches, " "))
     keyState.lastKeyTab = false
     ks.Redraw(prompt, buf)
    }
    keyState.lastKeyTab = false
   }
   // single tab behavior
   keyState.lastKeyTab = true
   switch len(matches) {

   case 0: // no matches do nothing
    fmt.Fprintf(ks.out, "\a")
   case 1:
    buf = []rune(matches[0]) // find the first relevant keyword
    buf = append(buf, ' ')   // add a space to the match
    ks.Redraw(prompt, buf)   // redraw the prompt + matched buffer + space
   // default:
   //  fmt.Fprintf(ks.out, "\r\n%s\r\n", strings.Join(matches, " "))
   //  ks.Redraw(prompt, buf)
   default:
    fmt.Fprintf(ks.out, "\a")
   }

  // support for backspace in terminal
  case keyBackspace, '\b':
   if len(buf) > 0 {
    buf = buf[:len(buf)-1]
    fmt.Fprint(ks.out, "\b \b")
   }
  case 3:
   return "", errors.New("interrupted")
```
