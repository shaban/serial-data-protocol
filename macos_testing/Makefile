.PHONY: all test-objcpp test-objcpp-zerocopy test-swift clean help

all: test-objcpp-zerocopy

# Test Objective-C++ bridge (SLOW - converts all data to ObjC objects)
test-objcpp:
	@echo "=== Building Objective-C++ Bridge Test (Object-based - SLOW) ==="
	cd objcpp_test && \
	clang++ -std=c++17 -ObjC++ -O3 -framework Foundation \
		test_objcpp.mm SDPAudioUnit.mm \
		../../testdata/audiounit_cpp/encode.cpp \
		../../testdata/audiounit_cpp/decode.cpp \
		-I../../testdata/audiounit_cpp \
		-o test_objcpp
	@echo ""
	@echo "=== Running Objective-C++ Bridge Test ==="
	cd objcpp_test && ./test_objcpp
	@echo ""

# Test Objective-C++ zero-copy bridge (FAST - keeps data in C++)
test-objcpp-zerocopy:
	@echo "=== Building Objective-C++ ZERO-COPY Bridge Test ==="
	cd objcpp_test && \
	clang++ -std=c++17 -ObjC++ -O3 -framework Foundation \
		test_objcpp_zerocopy.mm SDPAudioUnit_ZeroCopy.mm \
		../../testdata/audiounit_cpp/encode.cpp \
		../../testdata/audiounit_cpp/decode.cpp \
		-I../../testdata/audiounit_cpp \
		-o test_objcpp_zerocopy
	@echo ""
	@echo "=== Running Objective-C++ Zero-Copy Bridge Test ==="
	cd objcpp_test && ./test_objcpp_zerocopy
	@echo ""

# Test Swift with C bridge (most reliable)
test-swift:
	@echo "=== Building Swift C Bridge Test ==="
	cd swift_cpp_test && \
	swiftc -O test_swift_bridge.swift sdp_bridge.cpp \
		../../testdata/audiounit_cpp/encode.cpp \
		../../testdata/audiounit_cpp/decode.cpp \
		-I../../testdata/audiounit_cpp \
		-Xlinker -lc++ \
		-o test_swift_bridge
	@echo ""
	@echo "=== Running Swift C Bridge Test ==="
	cd swift_cpp_test && ./test_swift_bridge
	@echo ""

# Test Swift calling C++ directly (barebone - no wrapper)
test-swift-barebone:
	@echo "=== Building Swift Direct C++ Test (No Wrapper) ==="
	cd swift_cpp_test && \
	swiftc -O test_swift_cpp_direct.swift \
		../../testdata/audiounit_cpp/encode.cpp \
		../../testdata/audiounit_cpp/decode.cpp \
		-I. \
		-I../../testdata/audiounit_cpp \
		-cxx-interoperability-mode=default \
		-Xcc -std=c++17 \
		-Xlinker -lc++ \
		-o test_swift_cpp_direct
	@echo ""
	@echo "=== Running Swift Direct C++ Test ==="
	cd swift_cpp_test && ./test_swift_cpp_direct
	@echo ""

# Clean build artifacts
clean:
	rm -f objcpp_test/test_objcpp
	rm -f objcpp_test/test_objcpp_zerocopy
	rm -f swift_cpp_test/test_swift_bridge
	rm -f swift_cpp_test/test_swift_cpp_direct
	rm -rf swift_cpp_test/.build
	@echo "Cleaned build artifacts"

# Help
help:
	@echo "macOS C++ Interop Testing"
	@echo ""
	@echo "Available targets:"
	@echo "  all                   - Run zero-copy test (default)"
	@echo "  test-objcpp           - Test Objective-C++ bridge (object-based - SLOW)"
	@echo "  test-objcpp-zerocopy  - Test Objective-C++ zero-copy bridge (FAST)"
	@echo "  test-swift            - Test Swift 6 C++ interop (TODO: fix package)"
	@echo "  clean                 - Remove build artifacts"
	@echo ""
	@echo "Goal: Determine if C++ implementation can be used from Swift/ObjC"
	@echo "Success criteria: < 10% overhead vs pure C++ (29.3μs encode, 30.7μs decode)"
	@echo ""
	@echo "Note: Object-based bridge is ~1500% slower (converts all data to NSObjects)"
	@echo "      Zero-copy bridge keeps data in C++ and only exposes via methods"
