package main

type Builder interface {
	SetPartA()
	SetPartB()
	SetPartC()
	GetResult() Product
}

type Product struct {
	PartA string
	PartB string
	PartC string
}

type ConcreteBuilder struct {
	product Product
}

func (b *ConcreteBuilder) SetPartA() {
	b.product.PartA = "Part A"
}

func (b *ConcreteBuilder) SetPartB() {
	b.product.PartB = "Part B"
}

func (b *ConcreteBuilder) SetPartC() {
	b.product.PartC = "Part C"
}

func (b *ConcreteBuilder) GetResult() Product {
	return b.product
}

type Director struct {
	builder Builder
}

func (d *Director) Construct() {
	d.builder.SetPartA()
	d.builder.SetPartB()
	d.builder.SetPartC()
}

func main() {
	builder := &ConcreteBuilder{}
	director := &Director{builder: builder}
	director.Construct()
	product := builder.GetResult()
	println(product.PartA)
	println(product.PartB)
	println(product.PartC)
}
