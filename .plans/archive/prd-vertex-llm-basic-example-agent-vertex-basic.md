# PRD: Vertex LLM Basic Example (agent-vertex-basic)

## Overview
Add a new Vertex AI example by copying the existing `examples/agent-basic` and swapping the provider to Vertex, keeping the demo flow identical and minimal.

## Goals
- Provide a runnable Vertex AI example aligned with current agent-basic behavior.
- Demonstrate Vertex provider config using required env vars and ADC.

## Non-Goals
- No new tests, CI wiring, or additional example features.
- No new tooling or refactors outside the new example directory.

## Requirements

### Functional Requirements
- REQ-F-001: Create `examples/agent-vertex-basic` as a copy of `examples/agent-basic`.
- REQ-F-002: Update provider import and construction to `pkg/providers/vertex` with required `ProjectID`.
- REQ-F-003: Read config from env vars with minimal surface area, similar to agent-basic:
  - `VERTEX_PROJECT_ID` (required).
  - Optional: `VERTEX_LOCATION`, `VERTEX_MODEL` if present, otherwise defaults per provider.
- REQ-F-004: Preserve the same tools, prompts, output format, and run flow as in `examples/agent-basic/main.go`.

### Non-Functional Requirements
- REQ-NF-001: Keep example self-contained with no new dependencies.
- REQ-NF-002: Follow existing Go version guidance (1.26) and example patterns from `AGENTS.md`.

## Technical Considerations
- Vertex provider uses ADC; credentials are external (`GOOGLE_APPLICATION_CREDENTIALS` or workload identity). Do not embed credential loading in code.
- Vertex models and defaults per `pkg/providers/vertex/README.md`: default model `gemini-2.5-pro`, default location `us-central1`.
- Maintain agent config structure (timeouts, iteration count, tool registry) to keep behavior consistent with agent-basic.
- System prompt should remain in agent config; Vertex provider maps systemInstruction correctly.

## Acceptance Criteria
- [ ] New example exists at `examples/agent-vertex-basic` with a working `main.go`.
- [ ] Running the example with `VERTEX_PROJECT_ID` and valid ADC completes the same three demo sections as agent-basic.
- [ ] No changes required outside the example directory except any agreed README update.

## Out of Scope
- Example tests, CI changes, or additional docs beyond minimal README update.
- Additional Vertex features (streaming, extra tools, custom HTTP client).

## Open Questions
- Where should the “README update” live, given `examples/agent-basic` has no README?
- Should we add a link/mention in any top-level README or examples index?

Notes from sources used:
- `examples/agent-basic/main.go` for flow, tool setup, and output.
- `pkg/providers/vertex/README.md` for auth and config defaults.
- `AGENTS.md` for project conventions and example guidance.