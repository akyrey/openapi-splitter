package splitter

// SplitCallbacks extracts components/callbacks into individual files.
func SplitCallbacks(ctx *Context) error {
	return splitComponentMap(ctx, "callbacks", "callbacks")
}
