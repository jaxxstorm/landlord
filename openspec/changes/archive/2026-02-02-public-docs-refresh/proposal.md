## Why

The project needs public-facing documentation that clearly explains how Landlord works and how to try it locally. A coherent docs site (including API browsing) and a visible disclosure of OpenSpec/AI assistance will improve adoption, trust, and contributor onboarding.

## What Changes

- Update the root README to be a public overview with a quick example, a clear pointer to `docs/`, and a noticeable disclosure that the project was created with substantial OpenSpec and AI help.
- Create a `docs/` index README explaining Landlord’s architecture (compute, API, worker, database), its pluggable design, and a table of supported plugins by component.
- Add a local quickstart guide under `docs/`.
- Add detailed docs for compute providers, workflow providers (linking to public provider docs as needed), database types, and worker types.
- Ensure the docs are rendered via Docsify and avoid emojis.
- Add a Swagger UI page served as static docs content and styled to match the Docsify theme.

## Capabilities

### New Capabilities
- `public-docs`: Public-facing documentation set and Docsify site, including API browsing via Swagger UI.

### Modified Capabilities
- (none)

## Impact

- `README.md`
- `docs/` content structure and docsify configuration
- Swagger UI static assets or page included in docs
