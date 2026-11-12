package main

type GreetService struct{}

func (g *GreetService) Greet(name string) string {
	return "Hello " + name + "!"
}

func (g *GreetService) GreetMany(names []string) []string {
	greetings := make([]string, len(names))
	for i, name := range names {
		greetings[i] = "Hello " + name + "!"
	}
	return greetings
}
