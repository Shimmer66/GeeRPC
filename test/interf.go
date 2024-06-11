package main

import (
	"fmt"
	"time"
)

type Student struct {
	name string
	age  int
}
type Action interface {
	run(str string)
	sleep(str string)
	eat(str string) string
}

func (stu Student) run(s string) {
	fmt.Printf("%s has run %s mile", stu.name, s)
}

type IslandAnimal interface {
	move() error
	heartbeat() error
}

type Snake struct {
}

func (*Snake) move() error {
	return nil
}

func (*Snake) heartbeat() error {
	return nil
}

//func main() {
//	stu := Student{"xiao a", 12}
//	stu.run("12")
//
//}
func printDuration(seconds int) {
	duration := time.Duration(seconds) * time.Second
	fmt.Print(duration)
}

func main() {
	seconds := 10
	printDuration(seconds)
}
