# Micro-benchmarks for C encoder optimization research

CC = gcc
CFLAGS = -O3 -march=native -Wall -Wextra

TESTS = 0_writer_vs_buffer \
        1_layout_optimization \
        2_string_methods \
        3_array_optimization \
        4_struct_optimization

# Schema-based benchmarks (handwritten examples)
BENCHMARKS = bench_primitives bench_arrays bench_nested bench_complex

# Decoder benchmarks
DECODE_BENCHMARKS = decode_primitives_zerocopy \
                    decode_primitives_arena \
                    decode_audiounit_zerocopy

all: $(TESTS) $(BENCHMARKS) $(DECODE_BENCHMARKS)

%: %.c
	$(CC) $(CFLAGS) -o $@ $<

run: all
	@echo "================================="
	@echo "C Encoder Micro-Benchmark Suite"
	@echo "================================="
	@echo ""
	@for test in $(TESTS); do \
		echo "Running $$test..."; \
		./$$test; \
		echo ""; \
		echo "================================="; \
		echo ""; \
	done

run-benchmarks: $(BENCHMARKS)
	@echo "===================================="
	@echo "Schema-based Benchmark Results"
	@echo "===================================="
	@echo ""
	@for bench in $(BENCHMARKS); do \
		./$$bench; \
		echo ""; \
		echo "===================================="; \
		echo ""; \
	done

run-decode: $(DECODE_BENCHMARKS)
	@echo "===================================="
	@echo "Decoder Benchmark Results"
	@echo "===================================="
	@echo ""
	@for bench in $(DECODE_BENCHMARKS); do \
		./$$bench; \
		echo ""; \
		echo "===================================="; \
		echo ""; \
	done

clean:
	rm -f $(TESTS) $(BENCHMARKS) $(DECODE_BENCHMARKS)

.PHONY: all run run-benchmarks clean
