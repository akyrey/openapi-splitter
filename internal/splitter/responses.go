package splitter

// SplitResponses extracts components/responses into individual files.
func SplitResponses(ctx *Context) error {
	return splitComponentMap(ctx, "responses", "responses")
}
