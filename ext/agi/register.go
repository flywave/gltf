package agi

import (
	"github.com/flywave/gltf"
)

var registered = false

// RegisterExtensions registers the AGI extensions with the glTF library.
// This method must be called once at application's startup to register the extensions.
func RegisterExtensions() {
	if registered {
		return
	}

	registered = true

	// Register root level extensions
	gltf.RegisterExtension(ArticulationsExtensionName, UnmarshalAgiRootArticulations)
	gltf.RegisterExtension(StkMetadataExtensionName, UnmarshalAgiRootStkMetadata)
}

func init() {
	RegisterExtensions()
}
