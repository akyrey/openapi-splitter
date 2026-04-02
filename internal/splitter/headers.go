package splitter

// SplitHeaders extracts components/headers into individual files.
func SplitHeaders(ctx *Context) error {
	return splitComponentMap(ctx, "headers", "headers")
}
