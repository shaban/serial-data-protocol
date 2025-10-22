# Makefile for C++ Message Mode Tests and Benchmarks
# Part of Serial Data Protocol (SDP) test suite

CXX := c++
CXXFLAGS := -std=c++17 -Wall -Wextra -O2
LDFLAGS :=

# Generated source files
SOURCES := encode.cpp decode.cpp message_encode.cpp message_decode.cpp
HEADERS := types.hpp encode.hpp decode.hpp message_encode.hpp message_decode.hpp endian.hpp

# Test executables
TEST_MESSAGE := test_message_mode
TEST_CROSSLANG := test_crosslang
BENCH_MESSAGE := bench_message_mode

.PHONY: all test bench clean

all: test

# Run all tests
test: $(TEST_MESSAGE) $(TEST_CROSSLANG)
	@echo "=== Running C++ Message Mode Tests ==="
	@echo ""
	@./$(TEST_MESSAGE)
	@echo ""
	@./$(TEST_CROSSLANG)

# Run benchmarks
bench: $(BENCH_MESSAGE)
	@echo "=== Running C++ Message Mode Benchmarks ==="
	@echo ""
	@./$(BENCH_MESSAGE)

# Build test_message_mode
$(TEST_MESSAGE): test_message_mode.cpp $(SOURCES) $(HEADERS)
	@echo "Building $(TEST_MESSAGE)..."
	@$(CXX) $(CXXFLAGS) -o $@ test_message_mode.cpp $(SOURCES) $(LDFLAGS)

# Build test_crosslang
$(TEST_CROSSLANG): test_crosslang.cpp $(SOURCES) $(HEADERS)
	@echo "Building $(TEST_CROSSLANG)..."
	@$(CXX) $(CXXFLAGS) -o $@ test_crosslang.cpp $(SOURCES) $(LDFLAGS)

# Build bench_message_mode
$(BENCH_MESSAGE): bench_message_mode.cpp $(SOURCES) $(HEADERS)
	@echo "Building $(BENCH_MESSAGE)..."
	@$(CXX) $(CXXFLAGS) -o $@ bench_message_mode.cpp $(SOURCES) $(LDFLAGS)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(TEST_MESSAGE) $(TEST_CROSSLANG) $(BENCH_MESSAGE)
	@rm -f *.o

.DEFAULT_GOAL := test
