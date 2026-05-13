package main

import (
	"fmt"
)

type Product interface {
	RunApp() error
}

type Factory interface {
	Create() Product
}

type Phone struct {
	os string
}

type IosFactory struct{}

func (IosFactory) Create() Product {
	return &IosProduct{
		Phone: Phone{os: "ios"},
	}
}

type IosProduct struct {
	Phone
}

func (p *IosProduct) RunApp() error {
	fmt.Println(p.os + " app")
	return nil
}

type AndroidFactory struct{}

func (AndroidFactory) Create() Product {
	return &AndroidProduct{
		Phone: Phone{os: "android"},
	}
}

type AndroidProduct struct {
	Phone
}

func (p *AndroidProduct) RunApp() error {
	fmt.Println(p.os + " app")
	return nil
}

func run(f Factory) error {
	p := f.Create()
	return p.RunApp()
}

func main() {
	var f Factory
	f = AndroidFactory{}
	run(f)

	f = IosFactory{}
	run(f)
}
