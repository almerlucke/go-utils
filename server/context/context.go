package context

// Key to use as context key
type Key string

func (c Key) String() string {
	return "context key " + string(c)
}
