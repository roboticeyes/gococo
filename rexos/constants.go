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
	rexInspectionTargetType = "inspection.target"
	rexFileType             = "rex"

	// types for target child group nodes
	inspectionType = "inspection"
	activityType   = "activity"

	// types for tracks
	trackType = "track"
	routeType = "route"
)
