module github.com/shaban/serial-data-protocol/benchmarks

go 1.25.1

require (
	github.com/google/flatbuffers v23.5.26+incompatible
	github.com/shaban/serial-data-protocol v0.0.0-00010101000000-000000000000
	github.com/shaban/serial-data-protocol/testdata/generated/flatbuffers/go v0.0.0-00010101000000-000000000000
	github.com/shaban/serial-data-protocol/testdata/generated/protobuf/go v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.31.0
)

replace github.com/shaban/serial-data-protocol => ../../

replace github.com/shaban/serial-data-protocol/testdata/generated/protobuf/go => ../../testdata/generated/protobuf/go

replace github.com/shaban/serial-data-protocol/testdata/generated/flatbuffers/go => ../../testdata/generated/flatbuffers/go
