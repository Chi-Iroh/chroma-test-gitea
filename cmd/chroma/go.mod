module github.com/Chi-Iroh/chroma/v2/cmd/chroma

go 1.22

toolchain go1.24.5

replace github.com/Chi-Iroh/chroma-test-gitea => ../../

require (
	github.com/Chi-Iroh/chroma/v2 v2.0.0-20250729112305-f05d8ef28898
	github.com/alecthomas/kong v1.12.1
	github.com/mattn/go-colorable v0.1.14
	github.com/mattn/go-isatty v0.0.20
)

require (
	github.com/dlclark/regexp2 v1.11.5 // indirect
	golang.org/x/sys v0.29.0 // indirect
)
