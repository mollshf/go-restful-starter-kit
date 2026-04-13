package main

import "fmt"

type Foo struct {
	Bar string `json:"bar"`
}

func (f *Foo) Hello() {
	fmt.Println("Hello, World!")
}

func Sugoi(f *Foo) string {
	if f == nil {
		return "Foo is nil"
	}
	f.Bar = "Sugoi"
	return f.Bar
}

func main() {
	var foo *Foo

	fmt.Println(Sugoi(foo))
}
