package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"unicode"

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
			args := os.Args[1:]
			for _, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					if _, err := os.Stat(arg); err == nil {
						if err := RemoveImportSpace(arg); err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
					}
				}
			}
			os.Exit(RunGoimports(append([]string{"-local", modulePath}, args...)...))
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

func RemoveImportSpace(file string) error {
	in, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	writer := bytes.NewBuffer(nil)
	scanner := bufio.NewScanner(bytes.NewReader(in))
	inImport := false
	for scanner.Scan() {
		line := scanner.Text()
		if !inImport && line == "import (" {
			inImport = true
		}
		if inImport && line == ")" {
			inImport = false
		}
		if inImport && IsSpaceLine(line) {
			continue
		}
		_, err := fmt.Fprintln(writer, line)
		if err != nil {
			return err
		}
	}
	out := writer.Bytes()
	if !reflect.DeepEqual(in, out) {
		return ioutil.WriteFile(file, writer.Bytes(), 0644)
	}
	return nil
}

func IsSpaceLine(line string) bool {
	for _, c := range line {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}
