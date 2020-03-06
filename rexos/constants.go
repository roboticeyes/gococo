package rexos

var (
	apiFindByKey          = "/search/findByKey?key="
	apiFindByNameAndOwner = "/search/findByNameAndOwner?"
	apiFindByOwner        = "/search/findAllByOwner?owner="
	apiFindByParent       = "/search/findAllByParentReferenceAndCategory?parentReference="
	apiFindByUrn          = "/search/findByUrn?urn="
	apiAndCategory        = "&category="
	apiAndPage            = "&page="
)

const (
	rexSchemeV1             = "rexos.scheme.v1"
	rexFileType             = "rex"
)
