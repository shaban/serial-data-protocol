package integration

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// BenchmarkCrossLanguageSwiftEncode benchmarks Swift (C++ backend) encode performance via benchmark server
// NOTE: Swift package now uses C++ implementation via Swift/C++ interop
// TODO: Re-enable after verifying example_gen.go works with C++ backend
/*
func BenchmarkCrossLanguageSwiftEncode(b *testing.B) {
	schemaFile := filepath.Join("testdata", "primitives.sdp")
	outputDir := filepath.Join("testdata", "primitives", "swift")

	// Step 1: Regenerate with --mode=bench
	b.Log("Generating Swift code with benchmark mode...")
	genCmd := exec.Command("go", "run", "./cmd/sdp-gen",
		"-schema", schemaFile,
		"-output", outputDir,
		"-lang", "swift",
		"-mode", "bench")
	if out, err := genCmd.CombinedOutput(); err != nil {
		b.Fatalf("Generation failed: %v\nOutput: %s", err, out)
	}

	// Step 2: Build Swift helper
	b.Log("Building Swift helper...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	buildCmd := exec.CommandContext(ctx, "swift", "build", "-c", "release")
	buildCmd.Dir = outputDir
	if out, err := buildCmd.CombinedOutput(); err != nil {
		b.Fatalf("Swift build failed: %v\nOutput: %s", err, out)
	}

	// Step 3: Find the built binary (server name is based on package name: "swift-server")
	serverName := "swift-server"
	helperPath := filepath.Join(outputDir, ".build", "release", serverName)
	if _, err := os.Stat(helperPath); os.IsNotExist(err) {
		b.Fatalf("Swift server binary not found at %s", helperPath)
	}

	// Step 4: Run benchmark
	b.ResetTimer()
	results := runBenchmarkServer(b, helperPath, "--bench-server", "BENCH_ENCODE_ALLPRIMITIVES", "Swift")
	b.ReportMetric(results.opsPerSec, "ops/sec")
	b.ReportMetric(results.nsPerOp, "ns/op")
	b.Logf("Swift Encode: %.2f Mops/sec (%.2f ns/op)", results.opsPerSec/1e6, results.nsPerOp)
}
*/

// BenchmarkCrossLanguageRustEncode benchmarks Rust encode performance via benchmark server
func BenchmarkCrossLanguageRustEncode(b *testing.B) {
	schemaFile := filepath.Join("testdata", "primitives.sdp")
	outputDir := filepath.Join("testdata", "primitives", "rust")

	// Step 1: Regenerate with --mode=bench
	b.Log("Generating Rust code with benchmark mode...")
	genCmd := exec.Command("go", "run", "./cmd/sdp-gen",
		"-schema", schemaFile,
		"-output", outputDir,
		"-lang", "rust",
		"-mode", "bench")
	if out, err := genCmd.CombinedOutput(); err != nil {
		b.Fatalf("Generation failed: %v\nOutput: %s", err, out)
	}

	// Step 2: Build Rust server (server name is based on struct name: "all_primitives_server")
	b.Log("Building Rust server...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	serverName := "all_primitives_server"
	buildCmd := exec.CommandContext(ctx, "cargo", "build", "--release",
		"--example", serverName)
	buildCmd.Dir = outputDir
	if out, err := buildCmd.CombinedOutput(); err != nil {
		b.Fatalf("Rust build failed: %v\nOutput: %s", err, out)
	}

	// Step 3: Find the built binary
	helperPath := filepath.Join(outputDir, "target", "release", "examples", serverName)
	if _, err := os.Stat(helperPath); os.IsNotExist(err) {
		b.Fatalf("Rust server binary not found at %s", helperPath)
	}

	// Step 4: Run benchmark
	b.ResetTimer()
	results := runBenchmarkServer(b, helperPath, "--bench-server", "BENCH_ENCODE_ALL_PRIMITIVES", "Rust")
	b.ReportMetric(results.opsPerSec, "ops/sec")
	b.ReportMetric(results.nsPerOp, "ns/op")
	b.Logf("Rust Encode: %.2f Mops/sec (%.2f ns/op)", results.opsPerSec/1e6, results.nsPerOp)
}

type benchmarkResults struct {
	iterations uint64
	duration   time.Duration
	opsPerSec  float64
	nsPerOp    float64
}

// runBenchmarkServer spawns a helper in benchmark server mode and runs a timed benchmark
func runBenchmarkServer(tb testing.TB, helperPath, serverFlag, benchCmd, lang string) benchmarkResults {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start the helper in server mode
	cmd := exec.CommandContext(ctx, helperPath, serverFlag)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		tb.Fatalf("Failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		tb.Fatalf("Failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		tb.Fatalf("Failed to start %s server: %v", lang, err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Read the port from stdout
	scanner := bufio.NewScanner(stdout)
	var port string
	for scanner.Scan() {
		line := scanner.Text()
		tb.Logf("%s server output: %s", lang, line)
		if strings.HasPrefix(line, "BENCHPORT ") {
			port = strings.TrimPrefix(line, "BENCHPORT ")
			break
		}
	}

	if port == "" {
		// Read stderr for error messages
		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			tb.Logf("%s stderr: %s", lang, stderrScanner.Text())
		}
		tb.Fatalf("%s server did not print BENCHPORT", lang)
	}

	// Give server a moment to be ready
	time.Sleep(100 * time.Millisecond)

	// Connect to the server
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 5*time.Second)
	if err != nil {
		tb.Fatalf("Failed to connect to %s server: %v", lang, err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Send WARMUP command
	tb.Logf("Sending WARMUP to %s server...", lang)
	if _, err := conn.Write([]byte("WARMUP\n")); err != nil {
		tb.Fatalf("Failed to send WARMUP: %v", err)
	}
	response, err := reader.ReadString('\n')
	if err != nil {
		tb.Fatalf("Failed to read WARMUP response: %v", err)
	}
	if !strings.Contains(response, "OK") {
		tb.Fatalf("WARMUP failed: %s", response)
	}
	tb.Logf("Warmup complete")

	// Run iteration-based benchmark (eliminating timing overhead)
	// Use b.N for proper integration with Go's benchmarking framework
	iterations := uint64(1000000) // 1 million iterations for stable measurement
	tb.Logf("Running %s benchmark for %d iterations...", lang, iterations)

	start := time.Now()
	benchCommand := fmt.Sprintf("%s %d\n", benchCmd, iterations)
	if _, err := conn.Write([]byte(benchCommand)); err != nil {
		tb.Fatalf("Failed to send benchmark command: %v", err)
	}

	response, err = reader.ReadString('\n')
	elapsed := time.Since(start)
	if err != nil {
		tb.Fatalf("Failed to read benchmark response: %v", err)
	}

	// Verify OK response
	if !strings.Contains(response, "OK") {
		tb.Fatalf("Benchmark failed: %s", response)
	}

	// Calculate metrics
	actualDuration := elapsed.Seconds()
	opsPerSec := float64(iterations) / actualDuration
	nsPerOp := (actualDuration * 1e9) / float64(iterations)

	return benchmarkResults{
		iterations: iterations,
		duration:   elapsed,
		opsPerSec:  opsPerSec,
		nsPerOp:    nsPerOp,
	}
}
