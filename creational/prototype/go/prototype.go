package main

type Cloneable interface {
	Clone() Cloneable
}

type PrototpeManager struct {
	prototypes map[string]Cloneable
}

func (p *PrototpeManager) Register(name string, prototype Cloneable) {
	p.prototypes[name] = prototype
}

func (p *PrototpeManager) Unregister(name string) {
	delete(p.prototypes, name)
}

func (p *PrototpeManager) Get(name string) Cloneable {
	if prototype, ok := p.prototypes[name]; ok {
		return prototype.Clone()
	}
	return nil
}

func NewPrototpeManager() *PrototpeManager {
	return &PrototpeManager{
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
	manager := NewPrototpeManager()

	person := &Person{Name: "John", Age: 30}
	manager.Register("person", person)

	clonedPerson := manager.Get("person").(*Person)
	println(clonedPerson.Name) // Output: John
	println(clonedPerson.Age)  // Output: 30
}
