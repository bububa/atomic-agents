package systemprompt

// ContextProvider is an interface that defines the title and info of a context provider
type ContextProvider interface {
	Title() string
	Info() string
}
