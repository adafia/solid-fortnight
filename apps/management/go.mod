module github.com/adafia/solid-fortnight/apps/management

go 1.25.0

replace github.com/adafia/solid-fortnight/internal/config => ../../internal/config

replace github.com/adafia/solid-fortnight/internal/storage => ../../internal/storage

require (
	github.com/adafia/solid-fortnight/internal/config v0.0.0-00010101000000-000000000000
	github.com/adafia/solid-fortnight/internal/storage v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)

require (
	github.com/golang-migrate/migrate/v4 v4.19.1 // indirect
	github.com/lib/pq v1.11.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
