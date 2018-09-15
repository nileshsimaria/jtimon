package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	alias map[string]string
)

const (
	aliasFile = "alias.txt"
)

func mapInit() {
	f, err := os.Open(aliasFile)
	defer f.Close()
	if err != nil {
		return
	}
	alias = make(map[string]string)
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		tokens := strings.Split(line, ":")
		if len(tokens) != 2 {
			break
		}
		alias[strings.TrimSpace(tokens[1])] = strings.TrimSpace(tokens[0])
		fmt.Println(alias)
	}
}

func getAlias(input string) string {
	if alias == nil {
		return input
	}
	name, ok := alias[input]
	if !ok {
		return input
	} else {
		return name
	}
}
