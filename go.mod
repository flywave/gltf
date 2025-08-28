module github.com/flywave/gltf

require (
	github.com/flywave/go-draco v0.0.0-00010101000000-000000000000
	github.com/flywave/go-meshopt v0.0.0-00010101000000-000000000000
	github.com/flywave/go3d v0.0.0-20250619003741-cab1a6ea6de6
	github.com/flywave/webp v1.1.2
	github.com/go-test/deep v1.0.1
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

go 1.24

replace github.com/flywave/go-meshopt => ../go-meshopt

replace github.com/flywave/go-draco => ../go-draco
