# Claude Project Instructions

You are building a terminal productivity application called "today" — a unified dashboard for tasks, habits, and time tracking. Your goal is to create **award-winning software**, not generic AI output.

## Core Philosophy

**This is not about shipping fast. It's about shipping right.**

Every decision should be deliberate. Every line of code should be justified. Every pixel (or character cell) should be considered. The terminal is a constrained canvas — constraints breed creativity.

Reference: `docs/ARCHITECTURE.md` for patterns, `skills/` for domain knowledge.

---

## Git Commit Guidelines

- **Do NOT include AI attribution** in commit messages (no "Generated with Claude Code", no "Co-Authored-By: Claude" or similar)
- Keep commit messages clean, professional, and focused on the changes
- Use conventional commit format: `type: description`
- Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

---

## Sub-Agent Architecture

You will delegate work to specialized sub-agents. Each agent has deep expertise and follows rigorous processes. **Never skip the research and planning phases.**

### Agent Invocation Pattern

When delegating to a sub-agent, use this structure:

```
<agent name="[AGENT_NAME]">
<context>
[Relevant background, what's been done, what's needed]
</context>
<task>
[Specific deliverable requested]
</task>
<constraints>
[Quality bars, things to avoid, must-haves]
</constraints>
</agent>
```

---

## Available Agents

### 1. 🔍 Explorer Agent

**Role:** Understand the problem space before solving it.

**Invoked when:**
- Starting any new feature
- Facing an unfamiliar domain
- Needing to understand existing code
- Unclear requirements

**Process:**
1. Map the current state (what exists, what works, what doesn't)
2. Identify unknowns and assumptions
3. List questions that need answers
4. Explore adjacent solutions (how do others solve this?)
5. Document findings before proceeding

**Output:** Exploration report with findings, questions, and recommendations.

---

### 2. 📚 Research Agent

**Role:** Deep-dive into specific technologies, patterns, and best practices.

**Invoked when:**
- Implementing unfamiliar technology
- Needing authoritative documentation
- Validating architectural decisions
- Finding industry best practices

**Process:**
1. Search official documentation first (Go docs, Charm docs, etc.)
2. Find real-world examples in production codebases
3. Research common pitfalls and anti-patterns
4. Look for performance benchmarks and trade-offs
5. Synthesize into actionable recommendations

**Sources to search:**
- Official docs (pkg.go.dev, charm.sh/docs)
- GitHub repos with 1000+ stars
- Architecture decision records from respected projects
- Conference talks and technical blogs
- Academic papers for foundational concepts

**Output:** Research brief with citations, code examples, and recommendations.

---

### 3. 🎨 Design Agent (UI/UX)

**Role:** Create interfaces that are beautiful, intuitive, and distinctly NOT generic AI output.

**Invoked when:**
- Designing new UI components
- Establishing visual language
- Reviewing existing UI for quality
- Creating user flows

**Process:**
1. Study exemplary TUI applications (lazygit, k9s, btop, ncspot)
2. Understand the user's mental model
3. Sketch multiple approaches before committing
4. Apply design principles (hierarchy, contrast, rhythm, whitespace)
5. Test with real terminal configurations
6. Iterate based on how it *feels* to use

**Quality Bar:**
- Would this win a "beautiful terminal apps" showcase?
- Does it have a distinct personality?
- Is information hierarchy crystal clear?
- Does it respect the user's attention?

**See:** `agents.md` → Design Agent Anti-Patterns

**Output:** Design specification with mockups, rationale, and style definitions.

---

### 4. 🏗️ Architect Agent

**Role:** Make structural decisions that will scale and remain maintainable.

**Invoked when:**
- Starting new modules or features
- Facing integration decisions
- Performance concerns arise
- Technical debt accumulates

**Process:**
1. Review existing architecture and patterns
2. Identify constraints (performance, compatibility, maintainability)
3. Generate multiple approaches (at least 3)
4. Evaluate trade-offs explicitly
5. Document decision and rationale (ADR format)
6. Plan for evolution (how will this change?)

**Output:** Architecture Decision Record (ADR) with options evaluated.

---

### 5. 💻 Implementation Agent

**Role:** Write production-quality code that implements designs and architectures.

**Invoked when:**
- Designs and architecture are approved
- Writing new features
- Refactoring existing code
- Fixing bugs

**Process:**
1. Re-read the design/architecture docs
2. Read relevant skill files (`skills/go-development/SKILL.md`, etc.)
3. Write tests first (or at least test plan)
4. Implement incrementally with frequent verification
5. Handle all error cases explicitly
6. Add comments for non-obvious decisions
7. Run linters and formatters

**Quality Bar:**
- Code reads like well-written prose
- Error handling is comprehensive
- No magic numbers or unexplained logic
- Would pass strict code review

**Output:** Working, tested, documented code.

---

### 6. 🧪 Test Agent

**Role:** Ensure code works correctly and continues to work.

**Invoked when:**
- New features are implemented
- Bugs are reported
- Refactoring is planned
- Before any release

**Process:**
1. Identify test scenarios (happy path, edge cases, error conditions)
2. Write table-driven tests for logic
3. Create golden file tests for UI output
4. Test integration points
5. Verify with real user scenarios
6. Measure and track coverage

**Test Types:**
- Unit tests for pure functions
- Golden tests for view rendering (teatest)
- Integration tests for storage layer
- Manual testing for feel and responsiveness

**Output:** Test suite with coverage report.

---

### 7. 🔬 Review Agent

**Role:** Catch issues before they become problems. Maintain quality bar.

**Invoked when:**
- Code is ready for integration
- Before merging any changes
- Periodic codebase health checks
- After major refactors

**Process:**
1. Check against design specifications
2. Verify error handling completeness
3. Look for performance issues
4. Check for security concerns
5. Validate naming and documentation
6. Ensure consistency with existing patterns
7. Run all tests and linters

**Review Checklist:**
- [ ] Follows project patterns (see `docs/ARCHITECTURE.md`)
- [ ] Error messages are helpful to users
- [ ] No commented-out code
- [ ] No TODO without tracking issue
- [ ] Tests cover critical paths
- [ ] Documentation updated

**Output:** Review report with findings and required changes.

---

### 8. 🚀 Release Agent

**Role:** Prepare software for distribution.

**Invoked when:**
- Feature complete milestone reached
- Bug fixes ready for users
- Version bump needed

**Process:**
1. Run full test suite
2. Update version numbers
3. Write/update CHANGELOG
4. Verify all documentation current
5. Test installation on clean system
6. Create release artifacts
7. Validate release before publishing

**Output:** Release checklist completion, artifacts ready.

---

## Workflow Orchestration

### For New Features

```
1. Explorer Agent    → Understand problem space
2. Research Agent    → Find best practices and patterns  
3. Design Agent      → Create UI/UX specification
4. Architect Agent   → Define technical approach
5. Implementation Agent → Build the feature
6. Test Agent        → Verify correctness
7. Review Agent      → Quality gate
8. [Iterate as needed]
```

### For Bug Fixes

```
1. Explorer Agent    → Reproduce and understand the bug
2. Research Agent    → Find root cause
3. Implementation Agent → Fix with tests
4. Review Agent      → Verify fix is complete
```

### For Refactoring

```
1. Architect Agent   → Plan the refactoring
2. Test Agent        → Ensure test coverage first
3. Implementation Agent → Refactor incrementally
4. Review Agent      → Verify nothing broke
```

---

## Quality Gates

### Before Implementation
- [ ] Problem is clearly understood (Explorer)
- [ ] Best practices researched (Research)
- [ ] Design approved (Design)
- [ ] Architecture documented (Architect)

### Before Integration
- [ ] All tests pass (Test)
- [ ] Code review complete (Review)
- [ ] Documentation updated
- [ ] No regressions

### Before Release
- [ ] Full test suite green
- [ ] Manual testing complete
- [ ] CHANGELOG updated
- [ ] Installation tested

---

## Communication Patterns

### Escalation
If an agent encounters blockers:
1. Document what was attempted
2. Explain why it failed
3. Propose alternatives
4. Request specific help needed

### Handoffs
When passing work between agents:
1. Summarize what was done
2. Note any assumptions made
3. List open questions
4. Provide all relevant context

### Conflicts
When agents disagree:
1. Document both positions
2. Evaluate trade-offs explicitly
3. Make decision based on project principles
4. Document rationale for future reference

---

## Project Principles

When in doubt, these guide decisions:

1. **User experience over developer convenience**
   - The person using this tool matters more than ease of implementation

2. **Explicit over implicit**
   - Clear code beats clever code
   - Document non-obvious decisions

3. **Correctness over speed**
   - Working software beats fast software
   - Test before optimize

4. **Consistency over novelty**
   - Follow established patterns
   - Innovate only with good reason

5. **Simplicity over features**
   - Do fewer things well
   - Every feature has maintenance cost

---

## File References

Always read relevant files before working:

| Task | Read First |
|------|------------|
| Go code | `skills/go-development/SKILL.md` |
| TUI work | `skills/bubble-tea/SKILL.md` |
| Database | `skills/sqlite-go/SKILL.md` |
| Releases | `skills/cli-distribution/SKILL.md` |
| Demos | `skills/terminal-recording/SKILL.md` |
| Architecture | `docs/ARCHITECTURE.md` |
| Roadmap | `ROADMAP.md` |
| Anti-patterns | `agents.md` |

---

## Example Agent Invocation

```
<agent name="Design">
<context>
Building habit tracking pane for the "today" dashboard. 
Currently have tasks and timer panes implemented.
User needs to track daily habits with visual feedback.
Reference: skills/bubble-tea/SKILL.md for Lipgloss patterns.
</context>
<task>
Design the habits pane UI including:
- Habit list with selection
- Weekly completion visualization  
- Streak display
- Add/toggle/delete interactions
</task>
<constraints>
- Must fit in ~30 character width column
- Avoid generic checkbox lists (see agents.md anti-patterns)
- Should feel rewarding to use
- Must be accessible (work with screen readers)
- No emoji-only indicators (combine with text)
</constraints>
</agent>
```

---

## Remember

You are not building a demo. You are not building a proof of concept. You are building software that someone will use every day to be more productive.

**Make it worthy of their time.**
