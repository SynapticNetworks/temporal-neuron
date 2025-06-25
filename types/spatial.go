// types/spatial.go
package types

// =================================================================================
// SPATIAL DATA STRUCTURES
// =================================================================================

// Position3D defines spatial coordinates in 3D space for neural components
// This is used throughout the system for spatial organization, distance calculations,
// and anatomically-realistic neural network layouts.
//
// COORDINATE SYSTEM:
// - Units are typically in micrometers (μm) for cellular-level modeling
// - Origin (0,0,0) represents a reference point in the neural tissue
// - X, Y, Z axes follow standard 3D coordinate conventions
//
// BIOLOGICAL CONTEXT:
// Real neural networks have precise 3D organization:
// - Cortical layers (depth in Y or Z axis)
// - Columnar organization (clustering in X-Z plane)
// - Laminar structure (stratification along one axis)
// - Distance-dependent connectivity patterns
type Position3D struct {
	X float64 `json:"x"` // X-coordinate (lateral position)
	Y float64 `json:"y"` // Y-coordinate (anterior-posterior or depth)
	Z float64 `json:"z"` // Z-coordinate (vertical or layer position)
}

// BoundingBox defines a 3D rectangular region in space
// Used for spatial partitioning, collision detection, and region queries
type BoundingBox struct {
	Min Position3D `json:"min"` // Minimum corner of the box
	Max Position3D `json:"max"` // Maximum corner of the box
}

// Sphere defines a spherical region in 3D space
// Used for radial queries, diffusion modeling, and influence regions
type Sphere struct {
	Center Position3D `json:"center"` // Center point of the sphere
	Radius float64    `json:"radius"` // Radius of the sphere
}
