package splitter

// SplitParameters extracts components/parameters into individual files.
func SplitParameters(ctx *Context) error {
	return splitComponentMap(ctx, "parameters", "parameters")
}
