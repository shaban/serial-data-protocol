// swift-tools-version: 5.9
// Swift 5.9+ supports C++ interop (no need for Swift 6)

import PackageDescription

let package = Package(
    name: "SDPSwiftCppTest",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(
            name: "test-swift-cpp",
            targets: ["TestSwiftCpp"]
        ),
    ],
    targets: [
        // Test executable that uses C++ directly
        .executableTarget(
            name: "TestSwiftCpp",
            path: "Sources",
            sources: ["main.swift"],
            cxxSettings: [
                .unsafeFlags(["-std=c++17", "-O3"]),
                .headerSearchPath("../../testdata/audiounit_cpp")
            ],
            swiftSettings: [
                .interoperabilityMode(.Cxx),
                .unsafeFlags(["-O", "-I../../testdata/audiounit_cpp"])
            ],
            linkerSettings: [
                .unsafeFlags([
                    "../../testdata/audiounit_cpp/encode.cpp",
                    "../../testdata/audiounit_cpp/decode.cpp",
                    "-std=c++17",
                    "-O3"
                ])
            ]
        ),
    ],
    cxxLanguageStandard: .cxx17
)
