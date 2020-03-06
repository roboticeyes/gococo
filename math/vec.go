package math

// Vec3f defines a simple 3D vector
type Vec3f struct {
	X float32 `json:"x" example:"0.5"`
	Y float32 `json:"y" example:"0.5"`
	Z float32 `json:"z" example:"0.5"`
}

// Vec4f defines a simple 4D vector (used for rotation)
type Vec4f struct {
	X float32 `json:"x" example:"0.1"`
	Y float32 `json:"y" example:"0.1"`
	Z float32 `json:"z" example:"0.1"`
	W float32 `json:"w" example:"1.1"`
}
