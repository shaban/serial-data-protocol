# Project Constitution

**Version:** 1.0.0  
**Date:** October 18, 2025

## 1. Purpose

This constitution defines the governance, principles, and rules for the Serial Data Protocol project.

## 2. Core Documents

### 2.1 Canonical Documents

The project maintains **exactly five canonical documents**:

1. **DESIGN_SPEC.md** - Technical specification for version 1.0
2. **TESTING_STRATEGY.md** - Comprehensive testing approach
3. **DOCUMENTATION_GUIDELINES.md** - Documentation standards and practices
4. **IMPLEMENTATION_PLAN.md** - Step-by-step implementation tasks
5. **CONSTITUTION.md** - This document (project governance)

### 2.2 Document Authority

**These five documents are the sole source of truth** for project design, testing, documentation, implementation, and governance.

### 2.3 Prohibited Documents

The following document types are **explicitly prohibited** without amendment to this constitution:

- ❌ Overview documents
- ❌ Summary documents
- ❌ Quick reference documents
- ❌ Cheat sheets
- ❌ Roadmap documents (future versions)
- ❌ Architecture decision records (use inline in DESIGN_SPEC.md)
- ❌ Separate changelog (use git history)

**Rationale:** Duplicate information sources lead to inconsistency, confusion, and maintenance burden.

### 2.4 Allowed Documents

The following are **permitted** as they serve distinct purposes:

- ✅ `README.md` - Project introduction with links to canonical docs
- ✅ `LICENSE` - Legal licensing information
- ✅ `examples/*/README.md` - Example-specific instructions
- ✅ `docs/user-guide.md` - End-user guide (distinct from specification)
- ✅ `.gitignore`, `.editorconfig`, etc. - Tool configuration

## 3. Design Principles

### 3.1 Version Focus

**The project focuses exclusively on version 1.0.**

- ❌ No mentions of "v2", "future work", or "enhancements"
- ❌ No "deferred to later" features in specification
- ✅ Clear statement of current limitations in DESIGN_SPEC.md Section 11.5
- ✅ Workarounds documented for missing features

**Rationale:** Premature planning for future versions distracts from shipping v1.

### 3.2 Scope Discipline

**Version 1.0 scope is locked:**

- Schema format: Rust-like syntax
- Type system: Primitives, strings, arrays, structs
- Wire format: Little-endian, densely packed
- Encoder: C, with flexible field ordering
- Decoder: Go, with size limits
- Testing: Wire format based, no CGO in tests
- Target: Same-machine IPC only

**Any addition to scope requires constitution amendment.**

### 3.3 Simplicity Bias

**When in doubt, choose simplicity over features:**

- Fewer types > more types
- Fewer wire format features > more flexibility
- Fewer configuration options > more configurability

**Example:** Optional fields use array workaround (simple) rather than new `?` syntax (complex).

## 4. Decision Making

### 4.1 Design Decisions

**Process for design changes:**

1. **Identify issue** - What problem needs solving?
2. **Propose solution** - How to solve within v1 scope?
3. **Consider alternatives** - What are trade-offs?
4. **Update DESIGN_SPEC.md** - Document decision
5. **Update tests if needed** - Reflect in TESTING_STRATEGY.md

**No separate ADR (Architecture Decision Record) documents.**

### 4.2 Amendment Process

**To amend this constitution:**

1. Propose specific change to this document
2. Justify why change serves project goals
3. Update constitution with clear rationale
4. Update dependent documents if needed

**Threshold:** Unanimous agreement of active contributors.

## 5. Documentation Rules

### 5.1 Single Source of Truth

**Each piece of information exists in exactly one canonical document:**

| Information | Location |
|-------------|----------|
| Wire format spec | DESIGN_SPEC.md Section 6 |
| Schema syntax | DESIGN_SPEC.md Section 3 |
| Encoder API | DESIGN_SPEC.md Section 4 |
| Decoder API | DESIGN_SPEC.md Section 5 |
| Size limits | DESIGN_SPEC.md Section 5.5 |
| Testing approach | TESTING_STRATEGY.md |
| Documentation style | DOCUMENTATION_GUIDELINES.md |
| Implementation tasks | IMPLEMENTATION_PLAN.md |
| Project rules | CONSTITUTION.md (this document) |

### 5.2 Consistency Requirements

**When updating information:**

1. ✅ Update the single source of truth
2. ❌ Do NOT duplicate information elsewhere
3. ✅ Link to canonical source if needed

**Example:**
```markdown
❌ Bad: Repeat size limits in multiple documents
✅ Good: State in DESIGN_SPEC.md, link from elsewhere
```

### 5.3 Link Policy

**Internal links:**
- ✅ Link to specific sections: `[Size Limits](DESIGN_SPEC.md#55-size-limits)`
- ❌ Do not copy-paste content

**External links:**
- ✅ Link to official documentation (e.g., Go spec, Rust book)
- ❌ Link to blog posts or unofficial sources (they change)

## 6. Code Organization

### 6.1 Repository Structure

```
serial-data-protocol/
  cmd/
    sdp-gen/          # Generator binary
      main.go
  
  internal/
    parser/           # Schema parser
    validator/        # Schema validation
    generator/        # Code generation
      go/
      c/
      rust/
    wire/             # Wire format helpers
  
  testdata/           # Test schemas and fixtures (gitignored generated code)
  examples/           # Runnable examples
  
  DESIGN_SPEC.md             # Canonical spec
  TESTING_STRATEGY.md        # Canonical testing
  DOCUMENTATION_GUIDELINES.md # Canonical docs guide
  CONSTITUTION.md            # This document
  README.md                  # Project introduction only
  LICENSE                    # Legal
```

### 6.2 Generated Code

**Generated code is never checked into git (except in examples for demonstration).**

- `testdata/*/` - Gitignored, regenerated by TestMain
- `examples/*/generated/` - Included for demonstration, marked as generated

## 7. Version Control

### 7.1 Commit Messages

**Format:**
```
<type>: <summary>

<body>

<references>
```

**Types:**
- `spec:` - Changes to DESIGN_SPEC.md
- `test:` - Changes to TESTING_STRATEGY.md or test code
- `docs:` - Changes to DOCUMENTATION_GUIDELINES.md
- `gen:` - Changes to generator code
- `parser:` - Changes to schema parser
- `fix:` - Bug fixes
- `refactor:` - Code restructuring

**Example:**
```
spec: Add field reordering semantics

Document that scalar fields can be written in any order and are
reordered to schema definition order on commit. Update encoder
architecture section with tentative field storage design.

Refs: DESIGN_SPEC.md Section 4.1, 4.4
```

### 7.2 Branch Strategy

**Main branch:**
- `main` - Always deployable, all tests pass

**Development:**
- Feature branches: `feature/field-reordering`
- Fix branches: `fix/parser-error-messages`

**Merge requirements:**
- All tests pass
- Documentation updated
- Single source of truth maintained

## 8. Testing Requirements

### 8.1 Test Coverage

**Minimum coverage by component:**

- Parser: 90%
- Validator: 90%
- Generator: 85%
- Wire format helpers: 95%
- Integration tests: All happy paths + major error cases

**Enforcement:** CI fails if coverage drops below thresholds.

### 8.2 Test Organization

**Follow TESTING_STRATEGY.md exclusively:**

- Level 1: Unit tests (no generated code)
- Level 2: Generator tests
- Level 3: Integration tests (generated code)
- Level 4: Cross-language tests

**No ad-hoc test organization.**

## 9. Release Process

### 9.1 Version 1.0 Release Criteria

**Must complete:**

- ✅ All sections of DESIGN_SPEC.md implemented
- ✅ All tests in TESTING_STRATEGY.md passing
- ✅ Documentation in DOCUMENTATION_GUIDELINES.md followed
- ✅ Examples work and are documented
- ✅ Generator produces valid Go and C code
- ✅ Wire format specification complete and tested
- ✅ Cross-language tests pass (C→Go, Rust→Go)

**Tag:** `v1.0.0`

### 9.2 Post-1.0

**After 1.0 release:**

1. Update constitution to allow version 2.0 planning
2. Create new documents if needed for v2 design
3. Maintain v1.0 as stable branch

**Until then:** No v2 planning, no future work discussions in canonical documents.

## 10. Conflict Resolution

### 10.1 Document Conflicts

**If canonical documents contradict each other:**

1. Identify conflicting statements
2. Determine which is correct
3. Update incorrect document
4. Add clarifying cross-references

**Precedence (if unclear):**
1. CONSTITUTION.md (highest)
2. DESIGN_SPEC.md
3. TESTING_STRATEGY.md
4. DOCUMENTATION_GUIDELINES.md

### 10.2 Specification Ambiguity

**If DESIGN_SPEC.md is ambiguous:**

1. Clarify intended behavior
2. Update spec with precise language
3. Add example if helpful
4. Update tests to match

**Do NOT create separate clarification document.**

## 11. Enforcement

### 11.1 Pre-Commit Checks

**Automated checks:**
- Verify no prohibited documents added
- Verify generated code not committed (except examples)
- Verify links in canonical documents are valid

### 11.2 Pull Request Requirements

**Every PR must:**
- Update relevant canonical document(s)
- Include tests per TESTING_STRATEGY.md
- Follow DOCUMENTATION_GUIDELINES.md
- Maintain single source of truth

**Reviewer checklist:**
- [ ] Changes documented in canonical docs
- [ ] No duplicate information created
- [ ] Tests added/updated
- [ ] Constitution respected

## 12. Maintenance

### 12.1 Document Reviews

**Quarterly review of canonical documents:**
- Check for internal consistency
- Verify implementation matches specification
- Update examples if API changed
- Fix broken links

### 12.2 Cleanup

**If prohibited documents appear:**
1. Identify information source
2. Merge useful information into canonical docs
3. Delete prohibited document
4. Update links if any

## 13. Rationale

### 13.1 Why Four Documents?

**DESIGN_SPEC.md:**
- Technical truth about how system works
- Wire format, API, algorithms
- Implementation guide

**TESTING_STRATEGY.md:**
- How to verify correctness
- Test organization, fixtures, cross-language
- Quality assurance approach

**DOCUMENTATION_GUIDELINES.md:**
- How to write docs
- Style, examples, maintenance
- User-facing documentation

**CONSTITUTION.md:**
- Project governance
- Decision-making rules
- Document discipline

**Four is sufficient, more would fragment truth.**

### 13.2 Why No Version 2 Planning?

**Focus principle:** Planning future versions before shipping v1 is premature.

**Problems with future planning:**
- Distracts from completing v1
- Creates ambiguity ("is this v1 or v2?")
- Tempts feature creep
- Wastes effort on speculative design

**Better approach:** Ship v1, gather feedback, then plan v2.

### 13.3 Why Prohibit Summary Documents?

**Problem:** Summaries become outdated and contradict canonical docs.

**Example scenario:**
1. Create "Quick Reference" with wire format summary
2. Update wire format in DESIGN_SPEC.md
3. Forget to update Quick Reference
4. Users confused by conflicting information

**Solution:** No summaries. Link to canonical docs.

### 14. Amendment History

### Version 1.0.1 - October 18, 2025

- Added IMPLEMENTATION_PLAN.md as fifth canonical document
- Updated document count from four to five

### Version 1.0.0 - October 18, 2025

- Initial constitution
- Established four canonical documents
- Prohibited unauthorized documents
- Focused on version 1.0 exclusively

---

**End of Constitution**

**This document may only be amended by unanimous consent of active contributors.**

**Last amended:** October 18, 2025
