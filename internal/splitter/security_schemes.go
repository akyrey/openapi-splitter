package splitter

// SplitSecuritySchemes extracts components/securitySchemes into individual files.
func SplitSecuritySchemes(ctx *Context) error {
	return splitComponentMap(ctx, "securitySchemes", "securitySchemes")
}
