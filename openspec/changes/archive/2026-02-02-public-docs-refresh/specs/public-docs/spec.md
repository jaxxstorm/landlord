## ADDED Requirements

### Requirement: Public root README overview
The root README.md SHALL present a public-facing overview of Landlord with a concise description, a quick example of how it works, and a clear pointer to the docs site under `docs/`.

#### Scenario: Reader opens the root README
- **WHEN** a reader views README.md in the repository root
- **THEN** the README contains a brief overview, a quick example, and a link or reference to the docs in `docs/`

### Requirement: Visible OpenSpec and AI disclosure
The root README.md MUST include a noticeable disclosure that the project was created with substantial assistance from OpenSpec and AI.

#### Scenario: Reader scans README highlights
- **WHEN** a reader scans the primary sections of README.md
- **THEN** the OpenSpec and AI assistance disclosure is clearly visible without scrolling to a footer-only section

### Requirement: Docsify-rendered documentation site
The documentation in `docs/` SHALL be viewable via Docsify using a repository-local Docsify configuration.

#### Scenario: Docsify site is loaded
- **WHEN** a user opens the docs site entry point in `docs/`
- **THEN** Docsify renders the Markdown content into a navigable documentation site

### Requirement: Docs landing page with architecture and plugin matrix
The docs landing page in `docs/README.md` SHALL explain how Landlord works, describe the compute, API, worker, and database components, and include a table of supported plugins for each component, including the pluggable model.

#### Scenario: Reader opens docs landing page
- **WHEN** a reader opens `docs/README.md` via Docsify
- **THEN** the page includes component descriptions, pluggability notes, and a plugin support table

### Requirement: Local quickstart documentation
The docs SHALL include a quickstart guide that shows how to run Landlord locally.

#### Scenario: New user follows quickstart
- **WHEN** a new user navigates to the quickstart page
- **THEN** the page provides a minimal set of steps to run Landlord locally

### Requirement: Component detail pages
The docs SHALL include dedicated pages describing compute providers, workflow providers, database types, and worker types, linking to public provider documentation where applicable.

#### Scenario: Reader seeks provider details
- **WHEN** a reader navigates to the component detail pages
- **THEN** each component has its own page and external provider docs are referenced where needed

### Requirement: Swagger UI available within docs
The docs site SHALL provide a Swagger UI page that loads the OpenAPI specification generated into `docs/swagger.json` and can be accessed from Docsify navigation.

#### Scenario: Reader opens API documentation
- **WHEN** a reader selects the API documentation page in the docs navigation
- **THEN** Swagger UI loads using `docs/swagger.json` as the specification source

### Requirement: Consistent documentation styling
The Swagger UI page MUST apply styling consistent with the Docsify theme used by the docs site.

#### Scenario: Reader views Swagger UI
- **WHEN** the Swagger UI page is rendered
- **THEN** typography and colors align with the Docsify theme in use

### Requirement: Emoji-free documentation
Public documentation in README.md and `docs/` MUST avoid emojis.

#### Scenario: Documentation is reviewed
- **WHEN** a reviewer scans README.md and the docs pages
- **THEN** no emojis appear in headings, body text, or navigation
