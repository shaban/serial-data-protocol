// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "swift",
    products: [
        .library(
            name: "swift",
            targets: ["swift"]),
    ],
    targets: [
        .target(
            name: "swift",
            dependencies: []),
    ]
)
