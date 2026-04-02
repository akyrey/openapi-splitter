package splitter

// SplitLinks extracts components/links into individual files.
func SplitLinks(ctx *Context) error {
	return splitComponentMap(ctx, "links", "links")
}
