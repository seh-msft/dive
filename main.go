// Copyright (c) 2023, Microsoft Corporation, Sean Hinchee
// Licensed under the MIT License.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"
)

type Empty struct{}

var (
	byName      = flag.Bool("name", false, "Search for matching file name")
	mSize       = flag.Int("msize", 100, "Size of match buffer")
	winFlag     = flag.Bool("win", false, "Force using carriage technology")
	chatty      = flag.Bool("D", false, "Verbose logging output")
	maxFiles    = flag.Int("filemax", 500, "Maximum # of files to read at a time")
	noPlumb     = flag.Bool("N", false, "Do not include file:line: prefix")
	literal     = flag.Bool("literal", false, "Do not interpret regex, be literal")
	ms          = flag.Int("ms", 2, "Millisecond stagger on main goroutine")
	allDirs     = flag.Bool("a", false, "Don't skip directories like .git")
	allMimes    = flag.Bool("b", false, "Don't skip MIME typed files ≅ application/octet-stream")
	Nothing     = Empty{}
	sep         = byte('\n')
	fileAllowed chan Empty
	match       chan string
	expr        *regexp.Regexp
	raw         string
	noiseDirs   = []string{".git"}
)

// usage: dive 'some*thing' [dir ...]
func main() {
	flag.Parse()
	args := flag.Args()

	dirs := []string{"."} // cwd by default
	regex := ""
	out := bufio.NewWriter(os.Stdout)
	var wg sync.WaitGroup

	switch len(args) {
	case 0:
		fatal("fail: must call with an expression to match against")
	case 1:
		// Just an expression
		regex = args[0]

	default:
		// Expression and some # of dirs
		regex = args[0]
		dirs = args[1:]

	}

	var err error
	raw = regex
	if !*literal {
		expr, err = regexp.Compile(regex)
		efatal(err, "invalid regex provided")
	}

	if *winFlag {
		sep = '\r'
	}

	done := make(chan Empty)
	match = make(chan string, *mSize)
	fileAllowed = make(chan Empty, *maxFiles)
	for i := 0; i < *maxFiles; i++ {
		fileAllowed <- Nothing
	}

	for _, d := range dirs {
		wg.Add(1)
		go delve(d, &wg)
	}

	go func() {
		wg.Wait()
		done <- Nothing
		close(done)
	}()

SPIN:
	for {
		select {
		case s := <-match:
			chat("🔥 got a match!")
			fmt.Fprintln(out, s)
			out.Flush()

		case _, ok := <-done:
			if len(match) < 1 && !ok {
				break SPIN
			}

		default:
			time.Sleep(time.Duration(*ms) * time.Millisecond)
		}
	}
	out.Flush()
}

// Concurrenctly recursive directory diver
func delve(to string, wg *sync.WaitGroup) {
	var err error
	var entries []fs.DirEntry

	if !*allDirs {
		if slices.Contains(noiseDirs, to) {
			goto DONE
		}
	}

	chat("🤿 delving into", to)
	// If 'to is a non-dir file - jump straight to scrape
	if !isDir(to) {
		wg.Add(1)
		go scrape(to, wg)
		goto DONE
	}

	// If 'to' is a directory
	entries, err = os.ReadDir(to)
	if err != nil {
		// Ignore dir read fails - if we can't read, we can't read
		goto DONE
	}

DIRS:
	for _, entry := range entries {
		p := filepath.Join(to, entry.Name())
		if entry.IsDir() {
			// Delve deeper into directories
			wg.Add(1)
			go delve(p, wg)
		} else {
			// If matching file names - check name
			if *byName {
				if matches(p) {
					match <- p
					continue DIRS
				}
			} else {
				wg.Add(1)
				go scrape(p, wg)
			}
		}
	}

DONE:
	wg.Done()
}

func matches(s string) bool {
	if *literal {
		return strings.Contains(s, raw)
	}
	return expr.MatchString(s)
}

// Look through file lines to find a regex match
func scrape(to string, wg *sync.WaitGroup) {
	<-fileAllowed
	chat("🗃️ scraping at:", to)
	var r *bufio.Reader
	var ln int
	buf := make([]byte, 512) // Max # bytes considered by DetectContentType
	var mime string

	f, err := os.Open(to)
	if err != nil {
		// Ignore file read errors - if we can't read, we can't read
		chat("‽ file open err:", err)
		goto DONE
	}

	_, err = f.Read(buf)
	if err != nil {
		chat("‽ bytes read err:", "∈", to)
		goto DONE
	}

	mime = http.DetectContentType(buf)
	if !*allMimes && strings.Contains(mime, "octet-stream") {
		chat("\n\n\n\nMIME of", to, "→", mime, "\n\n\n")
		goto DONE
	}

	// Reset file reader after type detection
	f.Seek(0, 0)

	r = bufio.NewReader(f)
	ln = 1

	for ; ; ln++ {
		s, err := r.ReadString(sep)

		if err != nil && len(s) <= 0 {
			chat("‽ line read err:", err, "@ line:", ln, "∈", to)
			goto DONE
		}

		if strings.Contains(to, "config") {
			chat("📖 LINE:", s)
		}
		if matches(s) {
			s = strings.TrimSuffix(s, "\n")
			mstr := s
			if !*noPlumb {
				mstr = fmt.Sprintf("%s:%d: %s", to, ln, s)
			}
			chat("⚡ scrape MATCH @:", mstr)
			match <- mstr
			continue
		}
	}

DONE:
	f.Close()
	fileAllowed <- Nothing
	wg.Done()
}

/* utils */

func isDir(p string) bool {
	f, err := os.Open(p)
	if err != nil {
		chat("⚠️ isDir open fail →", err)
		return false
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		chat("⚠️ isDir stat fail →", err)
		return false
	}
	return stat.IsDir()
}

func chat(v ...any) {
	if !*chatty {
		return
	}
	fmt.Fprintln(os.Stderr, v...)
}

func efatal(err error, s ...any) {
	if err == nil {
		return
	}
	var msg []any = []any{"err:"}
	msg = append(msg, s...)
	msg = append(msg, "→", err.Error())
	fatal(msg...)
}

func fatal(s ...any) {
	fmt.Fprintln(os.Stderr, s...)
	os.Exit(1)
}
