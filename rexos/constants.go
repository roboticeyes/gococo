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
	RexSchemeV1 = "rexos.scheme.v1"
	RexFileType = "rex"

	ReferenceTypePortal = "portal"
	ReferenceTypeFile   = "file"
	ReferenceTypeRoot   = "root"
	ReferenceTypeGroup  = "group"

	ResourceTypeRexProject      = "RexProject"
	ResourceTypeRexReference    = "RexReference"
	ResourceTypeFileReference   = "FileReference"
	ResourceTypeProjectFile     = "ProjectFile"
	ResourceTypeGroup           = "GroupReference"
	ResourceTypeRootReference   = "RootReference"
	ResourceTypePortalReference = "PortalReference"
)
