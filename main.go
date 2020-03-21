package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/mod/modfile"
)

func main() {
	mod, err := GetGoEnv("GOMOD")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if mod != "" {
		content, err := ioutil.ReadFile(mod)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if modulePath := modfile.ModulePath(content); modulePath != "" {
			args := append([]string{"-local", modulePath}, os.Args[1:]...)
			os.Exit(RunGoimports(args...))
		}
	}
	os.Exit(RunGoimports(os.Args[1:]...))
}

func GetGoEnv(key string) (string, error) {
	output, err := exec.Command("go", "env").Output()
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		splits := strings.Split(scanner.Text(), "=")
		if len(splits) != 2 {
			return "", fmt.Errorf("invalid output: %s", output)
		}
		if key == splits[0] {
			return strings.Trim(splits[1], `"`), nil
		}
	}
	return "", nil
}

func RunGoimports(args ...string) int {
	cmd := exec.Command("goimports", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}
