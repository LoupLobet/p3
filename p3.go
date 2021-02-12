package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"log"
	"path/filepath"
)

const SHELL_PATH = "/bin/sh"
const SHELL_OPTION = "-c"
const DEFAULT_CONFIG = ".p3/default"

var qflag *bool

func main() {
	qflag = flag.Bool("q", false, "quiet")
	flag.Parse()

	if len(flag.Args()) == 0 {
		if EvalPath(DEFAULT_CONFIG, false) {
			RunConfig(DEFAULT_CONFIG)
		} else {
			Usage()
		}
	}
	for i := 0; i < len(flag.Args()); i++ {
		RunConfig(flag.Args()[i])
	}
}

func EvalConditions(conds string) bool {
	var token string
	var blank bool
	var escaped bool
	var negation bool

	for i := 0; i < len(conds); i++ {
		if !escaped {
			if conds[i] == '\\' {
				escaped = true
				continue
			} else if i == len(conds) - 1 {
				blank = true
				token += string(conds[i])
			} else if conds[i] == ' ' || conds[i] == '\t' {
				blank = true
				continue
			}
		}
		if blank {
			// new token reached
			if negation && len(token) < 1 {
				log.Fatal("negation on empty path")
			}
			if !EvalPath (token, negation) {
				return false
			}
			token = ""
			negation = false
		}
		if len(token) == 0 && conds[i] == '!' && !escaped {
			negation = true
		} else {
			token += string(conds[i])
		}
		blank = false
		escaped = false
	}
	return true
}

func EvalPath(path string, neg bool) bool {
	var bval bool

	// filepath.Glob() supports wildcards
	matches, err := filepath.Glob(path)
	if err != nil {
		log.Fatal(err)
	}
	bval = (len(matches) > 0)
	if neg {
		bval = !bval
	}
	return bval
}

func GetConditions(line string) (string, int) {
	for i := 0; i < len(line) - 1; i++ {
		if line[i] != '\\' && line[i + 1] == ':' {
			return line[:i + 1], i + 2
		}
	}
	return "", -1
}

func RemoveComment(line string) string {
	if len(line) > 0 && line[0] == '#' {
		return ""
	}
	for i := 0; i < len(line) - 1; i++ {
		if line[i] != '\\' && line[i + 1] == '#' {
			return line[:i]
		}
	}
	return line
}

func RunConfig(config string) {
	var line string
	var scnr *bufio.Scanner

	fd, err := os.Open(config)
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	scnr = bufio.NewScanner(fd)
	for scnr.Scan() {
		err := scnr.Err()
		if err != nil {
			log.Fatal(err)
		}
		line = scnr.Text()
		line = RemoveComment(line)
		if line == "" {
			continue
		}
		conditions, commandindex := GetConditions(line)
		if conditions == "" {
			log.Fatal("parsing: missing condition")
		}
		if EvalConditions(conditions) {
			stdout, err := RunShellCmd(scnr.Text()[commandindex:])
			if err != nil {
				log.Fatal(err)
			}
			if !*qflag {
				fmt.Printf("%s", stdout)
			}
		}
	}
}

func RunShellCmd(command string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(SHELL_PATH, SHELL_OPTION, command)
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.String(), err
}

func Usage() {
	fmt.Println("usage: p3 [-q] [config ...]")
	os.Exit(0)
}