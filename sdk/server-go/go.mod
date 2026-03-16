module github.com/adafia/solid-fortnight/sdk/server-go

go 1.25.0

replace github.com/adafia/solid-fortnight/internal/engine => ../../internal/engine

require (
	github.com/adafia/solid-fortnight/internal/engine v0.0.0-00010101000000-000000000000
)

require github.com/google/uuid v1.6.0 // indirect
