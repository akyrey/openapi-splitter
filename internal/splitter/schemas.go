package splitter

// SplitSchemas extracts components/schemas into individual files.
func SplitSchemas(ctx *Context) error {
	return splitComponentMap(ctx, "schemas", "schemas")
}
