## 1. Docsify scaffolding and navigation

- [x] 1.1 Add Docsify entry point `docs/index.html` with site name, theme, and load configuration
- [x] 1.2 Add `docs/_sidebar.md` (and `_navbar.md` if needed) to define docs navigation
- [x] 1.3 Ensure docsify config avoids emojis and references README-based landing page

## 2. Public-facing documentation content

- [x] 2.1 Rewrite root `README.md` as a public overview with quick example and docs pointer
- [x] 2.2 Add a visible OpenSpec/AI disclosure block in the root README
- [x] 2.3 Create `docs/README.md` explaining architecture, components, and pluggable model
- [x] 2.4 Add a plugin support matrix covering compute, workflow, database, and worker providers
- [x] 2.5 Add `docs/quickstart.md` with local run steps

## 3. Component detail pages

- [x] 3.1 Add or update docs for compute providers and link to provider-specific public docs
- [x] 3.2 Add or update docs for workflow providers and link to provider-specific public docs
- [x] 3.3 Add or update docs for database types
- [x] 3.4 Add or update docs for worker types

## 4. Swagger UI integration

- [x] 4.1 Add a Swagger UI page in `docs/` that loads `docs/swagger.json`
- [x] 4.2 Add CSS overrides so Swagger UI matches the Docsify theme
- [x] 4.3 Add navigation link to the Swagger UI page in Docsify sidebar

## 5. Validation

- [ ] 5.1 Verify Docsify renders `docs/` pages correctly in a browser
- [ ] 5.2 Verify Swagger UI loads and displays the OpenAPI spec
- [x] 5.3 Review docs for emoji-free content and disclosure visibility
