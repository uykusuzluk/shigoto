package main

import (
	"flag"
	"fmt"
	"strings"
)

var (
	tasker *string
)

const (
	template string = `
// auto-generated file by shigoto
// do not forget to fill out 'todo' marked places

package jobs

type %s struct {
	// TODO: Add necessary members
}

func (t *%s) New() (*shigoto.Runner, error) {
	return new (%s), nil
}

func (t *%s) Run() error {
	// TODO: Add code...
	return nil
}

func (t *%s) QName() string {
	return "%s"
}

func (t *%s) Identify() string {
	return "%s"
}

`
)

func main() {
	flag.Parse()
	lower := strings.ToLower(*tasker)

	a := fmt.Sprintf(template, *tasker, *tasker, *tasker, *tasker, *tasker, lower, *tasker, lower)
	fmt.Println(a)
}

func init() {
	tasker = flag.String("makeTask", "NewTask", "new task name for template generation")
}
