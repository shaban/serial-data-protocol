// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "SwiftResearch",
    platforms: [.macOS(.v13)],
    products: [
        .executable(
            name: "SwiftResearch",
            targets: ["SwiftResearch"]),
    ],
    targets: [
        .executableTarget(
            name: "SwiftResearch",
            swiftSettings: [
                .unsafeFlags(["-O", "-whole-module-optimization"]),
            ]),
    ]
)
