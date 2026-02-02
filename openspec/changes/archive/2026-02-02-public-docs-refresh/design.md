## Context

The root README is currently more developer- and implementation-oriented, while `docs/` contains many technical pages without a clear public entry point or navigation. There is no Docsify-based site, and the API documentation is primarily a raw OpenAPI spec (`docs/swagger.json` / `docs/swagger.yaml`) with references in the README. The change needs to reframe documentation for public consumption, provide a coherent docs structure, and surface API browsing in a themed docs site, while adding a noticeable disclosure about OpenSpec/AI assistance.

## Goals / Non-Goals

**Goals:**
- Provide a public-facing root `README.md` with a quick example, a pointer to `docs/`, and a clear OpenSpec/AI disclosure.
- Introduce a Docsify site for `docs/` with a main landing page that explains architecture, components, and the pluggable model.
- Include a component plugin matrix (compute, workflow, database, worker) on the docs landing page.
- Add a quickstart for local usage.
- Add detailed docs pages for compute, workflow providers (with links to provider docs), database types, and worker types.
- Serve Swagger UI as static docs content and keep it visually consistent with Docsify.
- Keep documentation text free of emojis.

**Non-Goals:**
- Changing API behavior, provider implementations, or database schemas.
- Rewriting all existing technical docs from scratch; this is a reorganization and augmentation for public clarity.
- Introducing new runtime dependencies outside the docs site (Docsify is client-side only).

## Decisions

1. **Docsify for static docs site**
   - **Decision:** Use Docsify with `docs/` as the site root.
   - **Alternatives:** MkDocs, Docusaurus, or a custom static site.
   - **Rationale:** Docsify is lightweight, requires no build step, and fits a simple static docs host for a Go project.

2. **`docs/README.md` as the public docs landing page**
   - **Decision:** Add a `docs/README.md` that explains architecture/components and presents the plugin matrix.
   - **Alternatives:** Keep the entry point in the root README or add a separate `docs/index.md`.
   - **Rationale:** Docsify uses `README.md` by default; this keeps navigation consistent and obvious.

3. **Curated navigation with `_sidebar.md` (and optional `_navbar.md`)**
   - **Decision:** Introduce a sidebar to highlight quickstart, components, and API docs.
   - **Alternatives:** Rely on inline links only.
   - **Rationale:** A visible nav reduces friction for new users and keeps the docs maintainable.

4. **Swagger UI embedded as a static docs page**
   - **Decision:** Add a docs page (e.g., `docs/api.html` or `docs/api.md` with embedded HTML) that loads Swagger UI and points at `docs/swagger.json`.
   - **Alternatives:** Redoc, or linking out to the raw OpenAPI spec only.
   - **Rationale:** Swagger UI is familiar to API users and aligns with existing OpenAPI output.

5. **Docsify theme alignment for Swagger UI**
   - **Decision:** Apply minimal CSS overrides to Swagger UI to match Docsify theme variables.
   - **Alternatives:** Accept default Swagger UI styling.
   - **Rationale:** Visual consistency improves the overall public presentation.

6. **Root README disclosure about OpenSpec/AI assistance**
   - **Decision:** Add a conspicuous disclosure block near the top of `README.md`.
   - **Alternatives:** Bury the disclosure in a footer or contributor docs.
   - **Rationale:** The request calls for visibility and transparency to readers.

## Risks / Trade-offs

- [Docsify requires client-side JS] → Mitigation: keep docs in Markdown in the repo so they remain readable without JS, and keep the root README informative.
- [Reorganized docs may break existing links] → Mitigation: keep existing files where possible, and add redirects or link aliases in Docsify if needed.
- [Swagger UI styling might not fully match Docsify] → Mitigation: scope CSS overrides to a minimal set and verify readability across light/dark variants if used.
- [Docs drift as features change] → Mitigation: add references to update steps (e.g., swagger regeneration) and keep the quickstart concise.

## Migration Plan

1. Inventory existing `docs/` content and identify canonical pages for components and providers.
2. Add Docsify scaffolding (`docs/index.html`, `_sidebar.md`, optional `_navbar.md`, and theme config).
3. Create `docs/README.md` with architecture overview and plugin matrix.
4. Add `docs/quickstart.md` and component-focused pages (compute, workflow providers, database, workers), linking to external provider docs as needed.
5. Add Swagger UI static page and theme-aligned CSS; wire it to `docs/swagger.json`.
6. Update root `README.md` with the new public overview, example, docs pointer, and OpenSpec/AI disclosure.
7. Validate locally by opening the Docsify site and ensuring the Swagger UI renders correctly.

## Open Questions

- Which compute, workflow, database, and worker providers should be listed as “supported” in the plugin matrix (source of truth: code, docs, or both)?
- Do we want to preserve existing doc filenames/paths or introduce a new layout with aliases?
- Where should the Swagger UI page live (HTML vs Markdown with embedded HTML), and do we prefer a docsify plugin approach?
