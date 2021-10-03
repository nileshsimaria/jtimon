package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Alias holds name mapping
type Alias struct {
	m map[string]string
}

// NewAlias is used to create new name mapping from given file
func NewAlias(file string) (*Alias, error) {
	f, err := os.Open(file)

	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := make(map[string]string)
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		tokens := strings.Split(line, ":")
		if len(tokens) != 2 {
			return nil, fmt.Errorf("alias mapping has syntax error for line (%s)", line)
		}
		m[strings.TrimSpace(tokens[1])] = strings.TrimSpace(tokens[0])
	}
	return &Alias{m: m}, nil
}

func getAlias(a *Alias, input string) string {
	if a == nil {
		return input
	}
	name, ok := a.m[input]
	if !ok {
		return input
	}
	return name
}
