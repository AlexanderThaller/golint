package main

import (
	"fmt"
	"io/ioutil"
	"opts"
	"os"
	"strings"
)

var version = "0.0.2"

// options
var showVersion = opts.Longflag("version",
	"display version information and exit")
var verbose = opts.Flag("v","verbose",
	"be verbose")

func main() {
	// Do the argument parsing
	opts.Parse()
	if *showVersion {
		ShowVersion()
		os.Exit(0)
	}
	for _, filename := range opts.Args {
		err := DoLint(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"golint: couldn't lint file: %s\n",
				filename)
		}
	}
}

// Show version information
func ShowVersion() {
	fmt.Printf("golint v%s\n",version)
}

var statelessLinters = []StatelessLinter {
	&LineLengthLint{},
	&TabsOnlyLint{},
	&TrailingWhitespaceLint{},
}

var statefulLinters = []StatefulLinter {
	&FilesizeLint{},
	&TrailingNewlineLint{},
}

type StatelessLinter interface {
	Lint(string) (string, bool)
}

type StatefulLinter interface {
	Lint(string, int) (string, bool)
	Reset()
	Done() (string, bool)
}

func DoLint(filename string) os.Error {
	// read in the file
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// prepare all the stateful linters
	for _, linter := range statefulLinters {
		linter.Reset()
	}
	// for each line
	lines := strings.Split(string(content), "\n", -1)
	for lineno, line := range lines {
		// run through the stateless linters
		for _, linter := range statelessLinters {
			msg, err := linter.Lint(line)
			if err {
				fmt.Printf("%s: L%d: %s\n",
					filename, lineno+1, msg)
			}
		}
		// run through the stateful linters
		for _, linter := range statefulLinters {
			msg, err := linter.Lint(line, lineno)
			if err {
				fmt.Printf("%s: %s\n",
					filename, msg)
			}
		}
	}
	// tell all the stateful linters we're done
	for _, linter := range statefulLinters {
		msg, err := linter.Done()
		if err {
			fmt.Printf("%s: %s\n", filename, msg)
		}
	}
	return nil
}
