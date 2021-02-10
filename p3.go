package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"log"
	"os/exec"
)

const SHELL_PATH = "/bin/sh"
const SHELL_OPTION = "-c"

// flags
var (
	oflag *bool
	sflag *bool
	vflag *bool
)

func main() {
	// parse command line
	oflag = flag.Bool("o", false, "output")
	sflag = flag.Bool("s", false, "symlink")
	vflag = flag.Bool("v", false, "verbose")
	flag.Parse()

	if len(flag.Args()) < 1 {
		Usage()
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
			} else if len(token) == 0 && conds[i] == '!' {
				negation = true
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
		}
		token += string(conds[i])
		blank = false
		escaped = false
	}
	return true
}

func EvalPath(path string, neg bool) bool {
	var bval bool

	_, err := os.Stat(path)
	bval = (err == nil)
	if neg {
		bval = !bval
	}
	return bval
}

func GetConditions(line string) (string, int) {
	for i := 0; i < len(line) - 1; i++ {
		if line[i] != '\\' && line[i + 1] == ':' {
			return line[:i], i + 2
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
			log.Println("parsing: missing condition")
		}
		if EvalConditions(conditions) {
			stdout, err := RunShellCmd(scnr.Text()[commandindex:])
			if err != nil {
				log.Fatal(err)
			}
			if *oflag {
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
	fmt.Println("usage: p3 [-ovs] [config ...]")
	os.Exit(0)
}