package main

import "fmt"

type Cloneable interface {
	Clone() Cloneable
}

type PrototypeManager struct {
	prototypes map[string]Cloneable
}

func (p *PrototypeManager) Register(name string, prototype Cloneable) {
	p.prototypes[name] = prototype
}

func (p *PrototypeManager) Unregister(name string) {
	delete(p.prototypes, name)
}

func (p *PrototypeManager) Get(name string) Cloneable {
	if prototype, ok := p.prototypes[name]; ok {
		return prototype.Clone()
	}
	return nil
}

func NewPrototypeManager() *PrototypeManager {
	return &PrototypeManager{
		prototypes: make(map[string]Cloneable),
	}
}

type Person struct {
	Name string
	Age  int
}

func (p *Person) Clone() Cloneable {
	return &Person{
		Name: p.Name,
		Age:  p.Age,
	}
}

func main() {
	manager := NewPrototypeManager()

	person := &Person{Name: "John", Age: 30}
	manager.Register("person", person)

	cloned := manager.Get("person").(*Person)
	fmt.Println(cloned.Name)
	fmt.Println(cloned.Age)
}
