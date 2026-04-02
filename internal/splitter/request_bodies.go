package splitter

// SplitRequestBodies extracts components/requestBodies into individual files.
func SplitRequestBodies(ctx *Context) error {
	return splitComponentMap(ctx, "requestBodies", "requestBodies")
}
