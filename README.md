# dive

A concurrent file tree explorer oriented towards source trees. 

Emits file names and lines by default in the style of [plumb(6)](http://man.postnix.pw/9front/6/plumb). 

Written in [Go](https://go.dev). 

## Build

	; go build

## Install

	; go install

## Usage

```
Usage of dive:
  -D    Verbose logging output
  -N    Do not include file:line: prefix
  -a    Don't skip directories like .git
  -b    Don't skip MIME typed files ≅ application/octet-stream
  -filemax int
        Maximum # of files to read at a time (default 500)
  -literal
        Do not interpret regex, be literal
  -ms int
        Millisecond stagger on main goroutine (default 2)
  -msize int
        Size of match buffer (default 100)
  -name
        Search for matching file name
  -win
        Force using carriage technology
```

## Examples

List all files from the current working directory in all sub-directories: 

```
» dive -name -a '.*'
README.md
go.mod
main.go
.git\HEAD
.git\config
.git\description
.git\hooks\applypatch-msg.sample
.git\hooks\commit-msg.sample
.git\hooks\fsmonitor-watchman.sample
.git\hooks\post-update.sample
.git\hooks\pre-applypatch.sample
.git\hooks\pre-commit.sample
.git\hooks\pre-merge-commit.sample
.git\hooks\pre-push.sample
.git\hooks\pre-rebase.sample
.git\hooks\pre-receive.sample
.git\hooks\prepare-commit-msg.sample
.git\hooks\push-to-checkout.sample
.git\hooks\update.sample
.git\info\exclude
»
```

Do a literal rather than a regex match in all sub-directories:

```
» dive -literal -a '[core]'
.git\config:1: [core]
»
```

Search across two directories in all sub-directories:

```
» dive -a '= [0-9]' .git ..\smolweb\.git
.git\config:2:  repositoryformatversion = 0
.git\hooks\pre-commit.sample:32:          LC_ALL=C tr -d '[ -~]\0' | wc -c) != 0
.git\hooks\pre-rebase.sample:20: if test "$#" = 2
.git\hooks\prepare-commit-msg.sample:33: #       if /^#/ && $first++ == 0' "$COMMIT_MSG_FILE" ;;
..\smolweb\.git\config:2:       repositoryformatversion = 0
.git\hooks\fsmonitor-watchman.sample:32: my $retry = 1;
.git\hooks\fsmonitor-watchman.sample:72:                "Falling back to scanning...\n" if $? != 0;
.git\hooks\fsmonitor-watchman.sample:130:                   "Falling back to scanning...\n" if $? != 0;
..\smolweb\.git\hooks\fsmonitor-watchman.sample:32: my $retry = 1;
..\smolweb\.git\hooks\fsmonitor-watchman.sample:72:             "Falling back to scanning...\n" if $? != 0;
..\smolweb\.git\hooks\fsmonitor-watchman.sample:129:                "Falling back to scanning...\n" if $? != 0;
..\smolweb\.git\hooks\pre-commit.sample:32:       LC_ALL=C tr -d '[ -~]\0' | wc -c) != 0
..\smolweb\.git\hooks\pre-rebase.sample:20: if test "$#" = 2
..\smolweb\.git\hooks\prepare-commit-msg.sample:33: #    if /^#/ && $first++ == 0' "$COMMIT_MSG_FILE" ;;
»
```

Find all asterisks in potential source files: 

```
» dive -literal '*' .
README.md:22: » dive -name -a '.*'
main.go:33:     expr        *regexp.Regexp
main.go:38: // usage: dive 'some*thing' [dir ...]
main.go:64:     if !*literal {
main.go:69:     if *winFlag {
main.go:74:     match = make(chan string, *mSize)
main.go:75:     fileAllowed = make(chan Empty, *maxFiles)
main.go:76:     for i := 0; i < *maxFiles; i++ {
main.go:105:                    time.Sleep(time.Duration(*ms) * time.Millisecond)
main.go:112: func delve(to string, wg *sync.WaitGroup) {
main.go:127:                    if !*all {
main.go:136:                    if *byName {
main.go:153:    if *literal {
main.go:160: func scrape(to string, wg *sync.WaitGroup) {
main.go:187:                    if !*noPlumb {
main.go:202: /* utils */
main.go:205:    if !*chatty {
»
```
