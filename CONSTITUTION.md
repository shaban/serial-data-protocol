# Project Constitution

**Version:** 2.0.0  
**Date:** October 18, 2025  
**Status:** Active (Release Candidate 0.2.0-rc1)

## 1. Purpose

This constitution defines the governance, principles, and rules for the Serial Data Protocol project. It captures the **design principles that helped us succeed** while updating documentation structure to reflect project maturity.

## 2. Core Documentation

### 2.1 Documentation Structure

The project maintains **focused, non-redundant documentation** organized in three tiers:

**Tier 1: Essential User Documentation**
1. **README.md** - Entry point: features, quickstart, installation
2. **DESIGN_SPEC.md** - Wire format specification (language-agnostic, authoritative)
3. **QUICK_REFERENCE.md** - API reference with practical examples
4. **CHANGELOG.md** - Version history and migration guides

**Tier 2: Performance & Safety**
5. **PERFORMANCE_ANALYSIS.md** - Benchmark methodology and detailed results
6. **BYTE_MODE_SAFETY.md** - Safety guide (when to use byte vs message mode)

**Tier 3: Contributor Documentation**
7. **LANGUAGE_IMPLEMENTATION_GUIDE.md** - Guide for implementing new language bindings
8. **DOCUMENTATION_GUIDELINES.md** - Documentation standards and practices
9. **TESTING_STRATEGY.md** - Cross-language testing approach

**Tier 4: Language-Specific Documentation**
- `go/README.md` - Go-specific setup, idioms, examples
- `c/README.md` - C-specific setup, build system, platform considerations
- `rust/README.md` - Rust-specific setup, idiomatic usage
- etc.

**Archive:**
- `archive/v0.1.0/` - Historical implementation plans for initial release
- `archive/v0.2.0-rc1/` - Historical RC planning documents
- `archive/analysis/` - Design exploration documents

### 2.2 Single Source of Truth Principle

**Each piece of information exists in exactly one authoritative location.**

| Information | Authoritative Location |
|-------------|----------------------|
| Wire format specification | DESIGN_SPEC.md Section 6 |
| RC features (optional, message mode, streaming) | DESIGN_SPEC.md Section 3 |
| Schema syntax | DESIGN_SPEC.md Section 2 |
| API examples | QUICK_REFERENCE.md |
| Performance numbers | PERFORMANCE_ANALYSIS.md + benchmarks/ |
| Safety guidelines | BYTE_MODE_SAFETY.md |
| Testing approach | TESTING_STRATEGY.md |
| Language porting guide | LANGUAGE_IMPLEMENTATION_GUIDE.md |
| Documentation style | DOCUMENTATION_GUIDELINES.md |
| Project governance | CONSTITUTION.md (this document) |

**Cross-references are encouraged, duplication is prohibited.**

### 2.3 Documentation Maintenance Rules

**When updating information:**
1. ✅ Update the single authoritative source
2. ✅ Link to authoritative source from related docs
3. ❌ Do NOT duplicate information
4. ❌ Do NOT create summaries that can go stale

**Example:**
```markdown
❌ Bad: Copy performance numbers into README.md
✅ Good: "See PERFORMANCE_ANALYSIS.md for detailed benchmarks"

❌ Bad: Repeat wire format details in QUICK_REFERENCE.md  
✅ Good: "Wire format: see DESIGN_SPEC.md Section 6"
```

### 2.4 New Document Policy

**Before creating a new document, ask:**
1. Does this information already exist? → Update existing doc
2. Is this temporary? → Put in archive/ or delete after use
3. Is this language-specific? → Put in `<language>/README.md`
4. Is this genuinely new and permanent? → Add to core documentation

**Prohibited document types:**
- ❌ Documents that duplicate existing information
- ❌ "Overview" or "summary" documents (use README.md)
- ❌ Architecture Decision Records (put rationale inline in DESIGN_SPEC.md)
- ❌ Multiple changelogs (one CHANGELOG.md only)

## 3. Design Principles

### 3.1 Focus on Current Version

**The project focuses on the current version being developed.**

✅ **Good practices:**
- Clear documentation of current features
- Honest limitations in README.md "When NOT to Use" section
- Workarounds documented for missing features
- Version history in CHANGELOG.md

❌ **Avoid:**
- Planning v2 features before v1 is released
- "Future work" sections in specs (use GitHub issues instead)
- Speculative features in documentation

**Rationale:** Focus on shipping the current version. Gather real-world feedback before planning the next version.

**Current status:** Version 0.2.0-rc1 (Release Candidate with optional fields, message mode, streaming I/O)

### 3.2 Simplicity Bias

**When in doubt, choose simplicity over features.**

Examples from our development:
- ✅ Fixed-width integers instead of variable-width (simpler, faster)
- ✅ Little-endian only (simpler, most platforms)
- ✅ No built-in compression (compose with gzip instead)
- ✅ Generated code with zero dependencies

**Decision framework:**
1. Can we solve this by composing existing features?
2. Does this add complexity to the wire format?
3. Does this add dependencies to generated code?
4. Can users implement this themselves if needed?

**Bias:** Prefer "no" to feature requests unless they solve a core problem that users cannot solve themselves.

### 3.3 Performance First

**Performance is a core feature, not an optimization.**

✅ **Maintained:**
- Verified benchmarks (see benchmarks/ directory)
- All performance claims backed by reproducible measurements
- Honest comparisons with Protocol Buffers and FlatBuffers
- Memory usage profiling

❌ **Reject:**
- Features that compromise performance without clear value
- "We'll optimize later" mindset
- Performance claims without benchmarks

**Benchmark discipline:**
- Use real-world data (not synthetic micro-benchmarks)
- Fair comparisons (same data, standard APIs, no tricks)
- Statistical confidence (multiple iterations)
- Document trade-offs honestly

### 3.4 Zero Dependencies for Generated Code

**Generated code must have zero runtime dependencies.**

✅ **Current state:**
- Go generated code: stdlib only
- C generated code: C11 stdlib only (planned)
- No compression libraries, no network libraries, no serialization frameworks

**Rationale:**
- Users compose SDP with their choice of compression/transport
- Reduces security surface area
- No version conflicts
- No supply chain risk

**Unix philosophy:** Provide interfaces (io.Reader/Writer), not implementations.

## 4. Decision Making

### 4.1 Design Changes

**Process for significant changes:**

1. **Identify problem** - What real-world problem needs solving?
2. **Check simplicity bias** - Can users solve this themselves?
3. **Propose solution** - How to solve within current principles?
4. **Document trade-offs** - What are the costs?
5. **Update DESIGN_SPEC.md** - Document decision with rationale inline
6. **Update tests** - Reflect in TESTING_STRATEGY.md
7. **Update benchmarks** - Measure performance impact

**No separate ADR (Architecture Decision Record) documents.** Rationale goes in DESIGN_SPEC.md.

### 4.2 Amendment to This Constitution

**Process:**
1. Propose specific change to CONSTITUTION.md
2. Justify why change serves project goals
3. Update constitution with clear rationale
4. Update amendment history (Section 15)

**Examples of valid amendments:**
- Updating documentation structure as project matures
- Adding new design principles that prove valuable
- Clarifying ambiguous governance rules

**Invalid amendments:**
- Removing simplicity bias
- Allowing documentation duplication
- Removing performance-first principle

## 5. Repository Structure

### 5.1 Multi-Language Organization

**Root structure:**
```
serial-data-protocol/
├── README.md                          # Entry point
├── DESIGN_SPEC.md                     # Wire format spec (language-agnostic)
├── QUICK_REFERENCE.md                 # API examples
├── CHANGELOG.md                       # Version history
├── PERFORMANCE_ANALYSIS.md            # Benchmarks
├── BYTE_MODE_SAFETY.md                # Safety guide
├── LANGUAGE_IMPLEMENTATION_GUIDE.md   # Porting guide
├── DOCUMENTATION_GUIDELINES.md        # Contributor docs
├── TESTING_STRATEGY.md                # Testing approach
├── CONSTITUTION.md                    # This document
├── LICENSE                            # MIT
│
├── go/                                # Go implementation
│   ├── README.md                      # Go-specific guide
│   ├── go.mod
│   ├── cmd/sdp-gen/                   # Generator
│   ├── internal/                      # Parser, validator, codegen
│   ├── integration_test.go
│   └── crossplatform_test.go
│
├── c/                                 # C implementation (planned)
│   ├── README.md                      # C-specific guide
│   ├── Makefile
│   └── cmd/sdp-gen-c/
│
├── rust/                              # Rust implementation (future)
│   ├── README.md
│   └── Cargo.toml
│
├── testdata/                          # Shared test schemas
│   ├── audiounit.sdp
│   ├── plugins.json
│   └── (generated code gitignored)
│
├── benchmarks/                        # Cross-language benchmarks
│   ├── README.md
│   ├── RESULTS.md
│   ├── MEMORY_ANALYSIS.md
│   └── (comparison tests)
│
└── archive/                           # Historical documentation
    ├── v0.1.0/                        # Initial implementation
    ├── v0.2.0-rc1/                    # RC planning
    └── analysis/                      # Design explorations
```

### 5.2 Language-Specific Documentation

**Each language implementation has its own README.md covering:**
- Installation and setup
- Build system (go.mod, Makefile, Cargo.toml, etc.)
- Language-specific API idioms
- Platform considerations
- Examples in that language
- Known limitations

**Wire format documentation stays in root DESIGN_SPEC.md (language-agnostic).**

### 5.3 Generated Code

**Generated code is never committed (except examples for demonstration):**
- `testdata/*/` - Gitignored, regenerated by tests
- `benchmarks/pb/`, `benchmarks/fb/` - Gitignored, regenerated by Makefile
- `go/examples/*/generated/` - May be committed for GitHub browsing

**.gitignore includes:**
```
*_generated.go
*_generated.c
*_generated.h
/testdata/primitives/
/testdata/audiounit/
# etc.
```

## 6. Version Control

### 6.1 Commit Messages

**Format:**
```
<type>: <summary>

<body>

<references>
```

**Types:**
- `spec:` - Changes to DESIGN_SPEC.md
- `test:` - Changes to tests or TESTING_STRATEGY.md
- `docs:` - Documentation updates
- `gen:` - Generator code changes
- `bench:` - Benchmark updates
- `fix:` - Bug fixes
- `refactor:` - Code restructuring
- `archive:` - Moving docs to archive/

**Example:**
```
spec: Add optional fields wire format

Document presence byte encoding for Option<T> fields.
Updated section 3.1 with wire format examples and decoder logic.

Refs: DESIGN_SPEC.md Section 3.1.2
```

### 6.2 Branch Strategy

**Simple flow:**
- `main` - Always releasable, all tests pass
- Feature branches: `feature/c-codegen`, `fix/parser-strings`
- Merge to main when: tests pass, docs updated, benchmarks run

## 7. Testing Requirements

**Follow TESTING_STRATEGY.md exclusively.**

**Minimum standards:**
- Parser/Validator: 90% coverage
- Generator: 85% coverage  
- Integration tests: All happy paths + major error cases
- Cross-language: Wire format compatibility

**Test organization:**
- Level 1: Unit tests (no generated code)
- Level 2: Generator tests
- Level 3: Integration tests (generated packages)
- Level 4: Cross-language tests

## 8. Release Process

### 8.1 Version Numbering

**Semantic versioning:** `MAJOR.MINOR.PATCH`

- **MAJOR:** Wire format breaking changes (rare!)
- **MINOR:** New features (optional fields, message mode, etc.)
- **PATCH:** Bug fixes, documentation

**Current:** 0.2.0-rc1 (Release Candidate)

### 8.2 Release Checklist

**Before release:**
- ✅ All tests pass (go test ./...)
- ✅ Benchmarks run successfully
- ✅ Documentation updated
- ✅ CHANGELOG.md has entry for version
- ✅ Performance claims verified
- ✅ Examples work

**Release tag:** `git tag v0.2.0-rc1`

## 9. Conflict Resolution

**If documentation contains contradictions:**

1. Identify conflicting statements
2. Determine which is correct (check implementation, tests, benchmarks)
3. Update incorrect documentation
4. Add clarifying note if needed

**Precedence (if truly ambiguous):**
1. CONSTITUTION.md (governance)
2. DESIGN_SPEC.md (wire format is authoritative)
3. Benchmarks (performance claims must match reality)
4. Other documentation

## 10. What Made This Project Successful

**Principles that helped us ship 0.2.0-rc1:**

✅ **Focus over features** - We shipped optional fields, message mode, and streaming instead of endlessly planning

✅ **Simplicity bias** - Fixed-width integers, no built-in compression, zero dependencies kept us fast

✅ **Performance first** - Fair benchmarks proved 6.1× faster encoding, 30% less RAM than Protocol Buffers

✅ **Single source of truth** - DESIGN_SPEC.md is authoritative for wire format, no contradictions

✅ **Test-driven** - 415 tests passing gave confidence to iterate quickly

✅ **Honest trade-offs** - "When NOT to Use" section builds trust, prevents misuse

**Principles to maintain:**
- Keep docs focused and non-redundant
- Verify all performance claims with benchmarks
- Maintain zero dependencies in generated code
- Document limitations honestly
- Prefer composition over built-in features

## 11. Amendment History

### Version 2.0.0 - October 18, 2025

**Major update after RC completion:**

- Updated documentation structure from "exactly five canonical documents" to three-tier system
- Added Tier 1 (user docs), Tier 2 (performance/safety), Tier 3 (contributor), Tier 4 (language-specific)
- Removed prohibition on quick reference and changelog (they proved valuable)
- Added "Performance First" and "Zero Dependencies" principles
- Updated repository structure for multi-language support
- Archived IMPLEMENTATION_PLAN.md and RC_IMPLEMENTATION_PLAN.md (complete)
- Simplified release process section
- Added "What Made This Project Successful" section capturing lessons learned

**Rationale:** Constitution v1.0 helped us stay focused during initial development. V2.0 updates rules to reflect project maturity while preserving core principles that worked.

### Version 1.0.1 - October 18, 2025

- Added IMPLEMENTATION_PLAN.md as fifth canonical document

### Version 1.0.0 - October 18, 2025

- Initial constitution
- Established governance rules
- Prohibited unauthorized documents
- Focused exclusively on version 1.0

---

**End of Constitution**

**This document captures the principles and practices that helped us succeed. Update it when those principles evolve, but don't discard what works.**

**Last amended:** October 18, 2025 (Version 2.0.0)
