module github.com/shaban/serial-data-protocol/benchmarks

go 1.25.1

require (
	github.com/google/flatbuffers v23.5.26+incompatible
	github.com/shaban/serial-data-protocol v0.0.0-00010101000000-000000000000
	github.com/shaban/serial-data-protocol/testdata/flatbuffers/go v0.0.0-00010101000000-000000000000
	github.com/shaban/serial-data-protocol/testdata/protobuf/go v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.31.0
)

replace github.com/shaban/serial-data-protocol => ../../

replace github.com/shaban/serial-data-protocol/testdata/protobuf/go => ../../testdata/protobuf/go

replace github.com/shaban/serial-data-protocol/testdata/flatbuffers/go => ../../testdata/flatbuffers/go
