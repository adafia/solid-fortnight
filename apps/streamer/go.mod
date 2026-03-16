module github.com/adafia/solid-fortnight/apps/streamer

go 1.25.0

replace github.com/adafia/solid-fortnight/internal/config => ../../internal/config

require (
	github.com/adafia/solid-fortnight/internal/config v0.0.0-00010101000000-000000000000
	github.com/redis/go-redis/v9 v9.7.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
