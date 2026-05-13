package main

type Product interface {
	runApp() error
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

func (p *IosProduct) runApp() error {
	println(p.os + " app")
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

func (p *AndroidProduct) runApp() error {
	println(p.os + " app")
	return nil
}

func run(f Factory) error {
	p := f.Create()
	return p.runApp()
}

func main() {

	var f Factory
	f = AndroidFactory{}
	run(f)

	f = IosFactory{}
	run(f)

}
