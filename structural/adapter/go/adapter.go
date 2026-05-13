package main

type Adaptee interface {
	SpecificRequest() string
}

type AdapteeImpl struct{}

func (a *AdapteeImpl) SpecificRequest() string {
	return "Specific Request"
}

func NewAdaptee() Adaptee {
	return &AdapteeImpl{}
}

type Target interface {
	Request() string
}

type Adapter struct {
	adaptee Adaptee
}

func (a *Adapter) Request() string {
	return a.adaptee.SpecificRequest()
}

func NewAdapter(adaptee Adaptee) Target {
	return &Adapter{adaptee: adaptee}
}

func main() {
	adaptee := NewAdaptee()
	adapter := NewAdapter(adaptee)

	println(adapter.Request()) // Output: Specific Request
}
