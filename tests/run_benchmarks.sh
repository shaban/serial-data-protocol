#!/usr/bin/env bash
# Run benchmarks for all languages

set -e

echo "========================================"
echo "SDP Benchmark Suite"
echo "========================================"
echo ""

# Parse arguments
LANG="${1:-all}"
COUNT="${2:-1}"

show_help() {
    echo "Usage: $0 [language] [count]"
    echo ""
    echo "Arguments:"
    echo "  language  Language to benchmark: go, cpp, rust, all (default: all)"
    echo "  count     Number of iterations for statistical reliability (default: 1)"
    echo ""
    echo "Examples:"
    echo "  $0              # Run all benchmarks once"
    echo "  $0 go           # Run Go benchmarks only"
    echo "  $0 all 10       # Run all benchmarks 10 times (for benchstat)"
    echo ""
    exit 0
}

if [ "$LANG" = "-h" ] || [ "$LANG" = "--help" ]; then
    show_help
fi

# Go benchmarks
run_go_benchmarks() {
    echo "[Go Benchmarks]"
    echo ""
    
    if [ ! -d "benchmarks" ]; then
        echo "❌ benchmarks/ directory not found"
        return 1
    fi
    
    cd benchmarks
    
    if [ "$COUNT" -gt 1 ]; then
        echo "Running $COUNT iterations for statistical analysis..."
        go test -bench=. -benchmem -count="$COUNT" -benchtime=1s | tee bench_results.txt
        echo ""
        echo "Results saved to benchmarks/bench_results.txt"
        echo "Analyze with: benchstat bench_results.txt"
    else
        go test -bench=. -benchmem -benchtime=3s
    fi
    
    cd ..
    echo ""
}

# C++ benchmarks
run_cpp_benchmarks() {
    echo "[C++ Benchmarks]"
    echo ""
    
    if [ -d "benchmarks/standalone/cpp" ]; then
        cd benchmarks/standalone/cpp
        if [ -f "Makefile" ]; then
            make bench
        else
            echo "  ℹ️  No Makefile found, skipping"
        fi
        cd ../../..
    else
        echo "  ℹ️  C++ benchmarks not yet implemented"
        echo "     See benchmarks/standalone/ for future implementation"
    fi
    echo ""
}

# Rust benchmarks
run_rust_benchmarks() {
    echo "[Rust Benchmarks]"
    echo ""
    
    if [ -d "benchmarks/standalone/rust" ]; then
        cd benchmarks/standalone/rust
        if [ -f "Cargo.toml" ]; then
            cargo bench
        else
            echo "  ℹ️  No Cargo.toml found, skipping"
        fi
        cd ../../..
    else
        echo "  ℹ️  Rust benchmarks not yet implemented"
        echo "     See benchmarks/standalone/ for future implementation"
    fi
    echo ""
}

# Main execution
case "$LANG" in
    go)
        run_go_benchmarks
        ;;
    cpp)
        run_cpp_benchmarks
        ;;
    rust)
        run_rust_benchmarks
        ;;
    all)
        run_go_benchmarks
        run_cpp_benchmarks
        run_rust_benchmarks
        ;;
    *)
        echo "❌ Unknown language: $LANG"
        echo ""
        show_help
        ;;
esac

echo "========================================"
echo "Benchmark Suite Complete"
echo "========================================"
echo ""
echo "Data source: testdata/data/plugins.json (115 KB, 62 plugins, 1,759 parameters)"
echo "See benchmarks/RESULTS.md for baseline performance"
echo "See benchmarks/BENCHMARK_DATA.md for methodology"
