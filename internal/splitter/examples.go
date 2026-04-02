package splitter

// SplitExamples extracts components/examples into individual files.
func SplitExamples(ctx *Context) error {
	return splitComponentMap(ctx, "examples", "examples")
}
