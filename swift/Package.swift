// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "SwiftResearch",
    platforms: [.macOS(.v13)],
    products: [
        .executable(
            name: "SwiftResearch",
            targets: ["SwiftResearch"]),
        .executable(
            name: "UnsafeBenchmarks",
            targets: ["UnsafeBenchmarks"]),
    ],
    targets: [
        .executableTarget(
            name: "SwiftResearch",
            path: "Sources/SwiftResearch",
            exclude: ["unsafe_benchmarks.swift"],
            sources: ["main.swift"],
            swiftSettings: [
                .unsafeFlags(["-O", "-whole-module-optimization"]),
            ]),
        .executableTarget(
            name: "UnsafeBenchmarks",
            path: "Sources/SwiftResearch",
            sources: ["unsafe_benchmarks.swift"],
            swiftSettings: [
                .unsafeFlags(["-O", "-whole-module-optimization"]),
            ]),
    ]
)
