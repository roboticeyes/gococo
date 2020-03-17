package rexos

// QueryFindByKey generates a FindByKey query
func QueryFindByKey(base, key string) string {
	return base + "/search/findByKey?key=" + key
}

// QueryFindByUrn generates a FindByUrn query
func QueryFindByUrn(base, urn string) string {
	return base + "/search/findByUrn?urn=" + urn
}

// QueryFindByParentReferenceAndCategory generates a query for getting a parent ref with category
func QueryFindByParentReferenceAndCategory(base, parent, category string) string {
	return base + "/search/findAllByParentReferenceAndCategory?parentReference=" + parent + "&category=" + category
}

// QueryGetPageAndSize generates a query with query parameters page and size
func QueryGetPageAndSize(base, page string, size string) string {
	return base + "?page=" + page + "&size=" + size
}
