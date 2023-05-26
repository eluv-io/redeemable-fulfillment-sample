package version

// Generates version information from git and stores in .../version-info.go

//go:generate echo go generate ./version/gen.go
//go:generate sh -c "\"../scripts/generate-version-info.sh\" > \"version-info.go\""
