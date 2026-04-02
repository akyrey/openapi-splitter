package splitter

// SplitPathItems extracts components/pathItems into individual files (OpenAPI 3.1).
func SplitPathItems(ctx *Context) error {
	return splitComponentMap(ctx, "pathItems", "pathItems")
}
