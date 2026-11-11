;lp# Code Review: Suggested Fixes

Suggestions from a review of the codebase, grouped by area. Line numbers refer to the code at
commit `b268547` and may have shifted since.

At review time `go test ./...` **failed**: `TestNew` panicked at `shell.go:151` because
`RunKeystrokeMode` type-asserted `sh.reader.(*os.File)` and the test passed a `bytes.Buffer`.

---

## Suggested order

1. Finish Step A for keylistener (see "Finishing Step A" below) and add the `LineReader`
   interface, so the tests pass again and you can write editor tests.
2. Write editor tests. They'll catch the fall-through and folder-path bugs.
3. Fix `FindLongestCommonPrefix`, the trie lowercasing, and the `findAllExeOSPath` early
   return. Each is small and easy to test.
4. Refactor `createPipeline` to return plain data, then switch execution to streaming before
   starting the pipes stage.

---

## 1. Making keylistener testable

### Why it's hard to test

`HandleInput` does three jobs at once:

1. **Terminal control.** `term.MakeRaw` needs a real file descriptor, so `in` has to be an
   `*os.File`.
2. **Reading input.** It reads runes in a loop.
3. **Line-editor logic.** Tab completion, double-tab, backspace, the buffer.

Only #3 really needs tests, but you can't reach it without #1. The fix is to separate them.

### Step A (smallest change): accept an `io.Reader` and only enter raw mode on a real TTY

```go
type KeyStrokeListener struct {
 in               *bufio.Reader // created ONCE, in the constructor
 raw              *os.File      // nil when input is not a file (e.g. in tests)
 out              io.Writer
 commandCompleter func(prefix string) []string
 fileCompleter    func(prefix string) ([]string, error)
}

func (ks *KeyStrokeListener) HandleInput(prompt string) (string, error) {
 if ks.raw != nil && term.IsTerminal(int(ks.raw.Fd())) {
  old, err := term.MakeRaw(int(ks.raw.Fd()))
  if err != nil {
   return "", err
  }
  defer term.Restore(int(ks.raw.Fd()), old)
 }
 // ...
}
```

**Bug this fixes:** `bufio.NewReader(ks.in)` used to be created on every `HandleInput` call.
With piped input such as `"echo a\recho b\r"`, the first reader buffers both lines and the
second line is lost when the function returns. Create the `bufio.Reader` once, in the
constructor.

#### Finishing Step A

The **struct field** is a `*bufio.Reader`, but the **constructor** should accept a plain
`io.Reader` and wrap it itself. That way `shell.New()` needs no changes (`sh.reader` is
already an `io.Reader`), and tests can pass `strings.NewReader(...)`.

The constructor must also set `raw`. Otherwise the real shell never enters raw mode and Tab
stops working:

```go
func NewKeyStrokeListener(r io.Reader, w io.Writer, cc func(string) []string, fc func(string) ([]string, error)) *KeyStrokeListener {
 ks := &KeyStrokeListener{
  in:               bufio.NewReader(r), // wrap once, here
  out:              w,
  commandCompleter: cc,
  fileCompleter:    fc,
 }
 // Only a real file can be a terminal; buffers and strings.Reader skip raw mode.
 if f, ok := r.(*os.File); ok {
  ks.raw = f
 }
 return ks
}
```

#### Example test

```go
func filterPrefix(words []string, prefix string) []string {
 var out []string
 for _, w := range words {
  if strings.HasPrefix(w, prefix) {
   out = append(out, w)
  }
 }
 return out
}

func TestTabCompletesSingleCommand(t *testing.T) {
 cmds := func(p string) []string { return filterPrefix([]string{"echo", "exit"}, p) }
 files := func(string) ([]string, error) { return nil, nil }
 var out bytes.Buffer
 ks := NewKeyStrokeListener(strings.NewReader("ec\t\r"), &out, cmds, files)

 got, err := ks.HandleInput("$ ")
 if err != nil {
  t.Fatal(err)
 }
 if want := "echo "; got != want {
  t.Errorf("got %q want %q", got, want)
 }
}
```

Use the same prompt as the shell (`"$ "`) so the output matches.

#### Expected follow-on test failure

Once `TestNew` in `shell_test.go` stops panicking, it will fail on its assertion. The listener
echoes each typed character to `out` (the `default:` case), so the output looks like
`"$ type exit\r\nexit is a shell builtin..."` instead of `"$ exit is a shell builtin\n"`.
That's correct terminal behaviour, so update the expected strings rather than the code. Better
still, assert only on the part after the echoed line, so the test checks what the command
printed and not how the editor drew the input.

### Step B (cleaner): pull the editor out as a pure state machine

```go
type Editor struct {
 buf     []rune
 lastTab bool
 cmds    func(string) []string
 files   func(string) ([]string, error)
}

type Effect struct {
 Bell   bool
 Redraw bool
 List   []string // candidates to print on double-tab
 Done   bool     // Enter pressed
}

func (e *Editor) Feed(r rune) Effect
func (e *Editor) Line() string
```

`KeyStrokeListener` then becomes a thin adapter: raw mode, read a rune, `Feed`, render the
`Effect`. Tests feed runes and check `Line()` and the `Effect`s. They never touch ANSI escape
codes, which is good because tests that match on `"\r\033[K..."` break easily.

### Step C (optional): a few end-to-end tests over a real pseudo-terminal

`github.com/creack/pty` lets you start the built binary under a pty and check the real
raw-mode behaviour. Keep these few; the editor tests do most of the work.

### Bugs in keylistener that the new tests will catch

- **File completion falls through into command completion.** After the
  `if commandExists(buf) {...}` block there's no `break`, so `cat ma<TAB>` completes the file
  and then also runs `commandCompleter("cat ma")`. That produces a second bell, and both
  branches share `lastKeyTab`.
- **Completing several files drops the folder.** `buf = []rune(command + LCP(files))` leaves
  out `folderPath`, so `cat dir/ma<TAB><TAB>` turns into `cat ma`.
- **Files and directories are told apart by extension.** `filepath.Ext(...) != ""` mishandles
  `Makefile` and `my.dir/`. `CompleteFileName` already appends `/` to directories, so check
  `strings.HasSuffix(name, "/")` instead.
- **`commandExists` doesn't do what its name says.** `idx+1 <= len(stringBuf)` is always true,
  so the function really means "contains a space and doesn't both start and end with a space".
  Its test is named `TestIsCommandFileCompletion`, has duplicate cases, and reports failures by
  index. Use `t.Run(tt.name, ...)`.
- Arrow keys and other escape sequences go straight into `buf`.
- Backspace doesn't reset `lastKeyTab`.

---

## 2. Architecture: separate parse → plan → execute

Keep pure logic away from I/O and the OS:

```
string ──ParseLine──▶ tokens ──Plan──▶ []Command{Name, Args, Redirects} ──Exec──▶ side effects
        (pure ✅)           (pure, should be)                         (inject I/O + runner)
```

### `createPipeline` should return plain data

Right now it takes a `Shell` **by value** and stores bound method values
(`sh.noRedirection` and so on) in a map indexed 0/1/2. Have it return something like this:

```go
type Redirect struct {
 Fd     int
 Path   string
 Append bool
}

type Command struct {
 Name      string
 Args      []string
 Redirects []Redirect
}
```

That's easy to table-test, like the parser tests. It also exposes a real bug: **the redirect
flags are never reset.** In `echo hi > a.txt 2> b.txt`, every token after `>` is treated as the
output filename. You end up with `stdOutFile = "b.txt"` and stderr is never redirected.

### Stream output instead of buffering it

`result{stdOut bytes.Buffer}` collects all output before writing anything. That breaks
long-running or interactive programs, and `execCmd.Stdin` is never set. Before running a
command, work out its `io.Reader` and `io.Writer`s: redirection means opening a file and using
it as the writer. That replaces `noRedirection`, `redirectStdOut`, `redirectStdError` and the
`output` map with about 15 lines. You'll need this anyway for the `|` pipelines stage.

### Give builtins a context and an exit code

```go
// ctx has Stdin/Stdout/Stderr, Cwd, Env
type Builtin func(ctx *ExecContext, args []string) int
```

That makes every builtin testable with `bytes.Buffer`. It also gets `os.Exit` out of
`exitHandler` (have it return a sentinel error or code that the main loop handles), which
currently makes `exit` impossible to test.

### Add a separate stderr writer

Errors currently go to the stdout writer.

### Make `New()` lighter

It reads the cwd and `PATH`, and scans every PATH directory, every time it's called, including
in every test. Accept a config instead (`Stdin`, `Stdout`, `Stderr`, `Env`, a line reader, a
command runner) and have `main` supply the real ones.

### Have `Shell` depend on a small interface

```go
type LineReader interface {
 ReadLine(prompt string) (string, error)
}
```

In production that's `KeyStrokeListener`. In tests it's a fake that returns preset lines.

---

## 3. Other bugs

| Where | Issue |
| --- | --- |
| `completion/command.go:90` `FindLongestCommonPrefix` | It doesn't stop at the first mismatch: `["abcd","abxd"]` returns `"abd"`. It also mixes byte indexes with rune iteration and sorts the caller's slice in place. |
| `completion/command.go:33` `Trie.Insert` | It lowercases words, but `Search` and `Prefixes` don't. `README.md` completes as `readme.md`, and typing `RE<TAB>` finds nothing. |
| `completion/path.go:16` | `TrimPrefix(dir, "/")` turns absolute paths into relative ones: `cat /tmp/fo<TAB>` searches `./tmp`. |
| `shell.go:204` `findAllExeOSPath` | It returns on the first unreadable PATH entry instead of `continue`. One stale PATH directory hides every executable after it. |
| `shell.go:251` | `execCmd.ProcessState` is nil when the process fails to start, so this line will panic. Exit codes from `ExitError` are never recorded. |
| `handlers.go:72,76` `cd` | The error from `os.Chdir(home)` is ignored. Errors other than not-exist (not a directory, permission denied) are silently swallowed. `~/foo` isn't supported. |
| `handlers.go` `type` | With several arguments it returns after the first one that isn't a builtin. |

---

## 4. Tests and test hygiene

- **Tests write into the repo.** `TestCreatePathAndFile` writes `testdata/t1/test1.txt`, and
  `TestCreatePipeline` writes `test.txt`. Use `t.TempDir()` and `t.Chdir()` (available since
  Go 1.24).
- **`TestCreatePipeline` depends on the system `cat`** and its exact error text. Once execution
  goes through an injectable runner, use a fake.
- **`path_test.go` is empty.** `CompleteFileName` is easy to test once it takes a base
  directory (or an `fs.FS`, so tests can use `fstest.MapFS`) instead of calling `os.Getwd()`.
- Use `t.Run(name, ...)` everywhere. The parser tests already do this well.

---

## 5. Cleanup

- Remove dead state: `Shell.Paths`/`ParsePath` (you use `exec.LookPath`), `WorkingDir` (never
  updated by `cd`), `IsExecAll`, the `tokenizer` stub, and the `fmt` workaround in `main.go`.
- `interceptor` exists to add a missing newline after builtins whose output doesn't end in one.
  Have each builtin print its own `\n` and delete the interceptor. Its `Write` also loops over
  every byte just to find the last one.
- `fmt.Errorf("%v", err)` adds nothing; use `%w` or return `err` directly.
- `output.txt` and `issues.md` are committed. Consider `.gitignore`.
- Replacing the trie in `CompleteFileName` with a `strings.HasPrefix` filter would be simpler.

---

## What's already good

- `ParseLine` is a clear state machine with thorough table tests. Use it as the model for the
  rest.
- Injecting the completers as functions into the listener was the right call. Steps A and B
  build on it.
- The packages are split sensibly.
