## Context

The current `landlord` application uses `alecthomas/kong` for configuration management. Kong handles CLI flag parsing and environment variable loading via struct tags, but lacks:
- Multi-format configuration file support (YAML, JSON, TOML)
- Structured configuration precedence (CLI > env > file > defaults)
- Hot-reload awareness and configuration watching
- Direct integration with the broader Go CLI ecosystem

The application currently:
- Loads config exclusively from environment variables and CLI flags
- Uses kong struct tags for validation and metadata
- Has a centralized `internal/config/Config` struct

We need to preserve all current functionality (env vars, CLI flags) while adding file-based configuration support. This is a cross-cutting change affecting the main entry point and the configuration package.

## Goals / Non-Goals

**Goals:**
- Replace kong with cobra/viper while maintaining environment variable and CLI flag support
- Add YAML and JSON configuration file support
- Establish clear precedence: CLI flags > environment variables > config file > defaults
- Preserve existing environment variable names and meanings (backward compatibility)
- Provide seamless struct binding via viper (e.g., `viper.Unmarshal(&cfg)`)
- Update documentation with configuration examples
- Maintain 100% test coverage for configuration loading logic
- Allow ability to easily add configuration for compute providers, workflow providers and database provider be ensuring they are configured in separate files within the configuration directories
- Allow abbilty to define a compute provider, workflow provider or database provider on startup

**Non-Goals:**
- Hot-reload of configuration during runtime (deferred to future iteration)
- TOML, INI, or other format support (start with YAML and JSON only)
- Interactive CLI wizards or advanced cobra subcommand structure (basic single command for now)
- Configuration validation refactoring (keep existing validation logic as-is)

## Decisions

### Decision 1: Use Cobra for CLI Framework
**Choice**: Adopt `spf13/cobra` for command parsing and help generation.

**Rationale**: 
- Cobra is the standard in the Go ecosystem (kubectl, Docker, Hugo)
- Seamless integration with viper for flag/config precedence
- Better help text and subcommand support
- Community-maintained with excellent documentation

**Alternatives Considered**:
- Urfave/cli: Simpler but less ecosystem integration
- Pflag alone: Low-level, doesn't provide subcommand structure
- Keep kong: Would need to add viper separately, less integrated

### Decision 2: Use Viper for Configuration Management
**Choice**: Adopt `spf13/viper` for unified config file + env + flag handling.

**Rationale**:
- Single library handles all configuration sources with proper precedence
- Struct binding via `Unmarshal()` reduces boilerplate
- Supports YAML, JSON, TOML, HCL, INI, and environment variables out-of-box
- Works seamlessly with cobra flags

**Alternatives Considered**:
- Manual precedence logic: Error-prone and complex to maintain
- Separate file parsing library: Would duplicate viper's functionality

### Decision 3: Configuration Precedence Order
**Choice**: CLI flags > environment variables > config file > defaults

**Rationale**:
- Matches user expectations (what I type now overrides config files)
- Standard pattern in infrastructure tooling (kubectl, docker, terraform)
- Environment variables can be managed by operators, config files by admins

**Implementation**:
- Cobra flags have highest priority (checked first by viper)
- Viper automatically handles env var overrides via `BindEnv()`
- Config file loaded by `ReadConfig()` with defaults as fallback

### Decision 4: Configuration File Locations and Naming
**Choice**: 
- Look for `config.yaml` or `config.json` in: current dir, `/etc/landlord/`, and `$XDG_CONFIG_HOME/landlord/`
- Support `--config` flag to specify explicit file path
- Environment variable `LANDLORD_CONFIG` can override file location

**Rationale**:
- Current dir: convenient for development and Docker containers
- `/etc/landlord/`: Linux standard for system configurations
- XDG: respects freedesktop standards on user systems
- Explicit override: flexibility for custom deployments

### Decision 5: Struct Tags for Viper Binding
**Choice**: Use `mapstructure` tags on config structs (viper's default tagging).

**Rationale**:
- Viper uses mapstructure for unmarshaling by default
- Cleaner than kong's inline validation
- Decouples config definition from command parsing

**Example**:
```go
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
    HTTP     HTTPConfig     `mapstructure:"http"`
}
```

YAML example:
```yaml
database:
  host: localhost
  port: 5432
http:
  host: 0.0.0.0
  port: 8080
```

### Decision 6: Preserve Environment Variable Names
**Choice**: Keep existing env var naming (`DATABASE_HOST`, `HTTP_PORT`, etc.) unchanged.

**Rationale**:
- Zero disruption to existing deployments
- Current scripts and CI/CD systems continue working
- Operators familiar with current names

**Implementation**:
```go
viper.BindEnv("database.host", "DATABASE_HOST")
viper.BindEnv("http.port", "HTTP_PORT")
// etc.
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| **Breaking change in imports** - Code importing kong types will fail | Refactor main.go and config.go to remove kong dependencies; comprehensive testing of config loading path |
| **Viper's complexity** - More flexible than needed, potential learning curve | Document viper usage clearly; provide example configs; keep config struct simple |
| **Config file precedence confusion** - Users unclear which setting wins | Document precedence clearly in README; log effective configuration at startup with source (CLI vs env vs file) |
| **Increased dependency count** - Adding cobra and viper (slight binary size impact) | Benefit of multi-format support outweighs ~5MB binary size increase |
| **Validation moved away from struct tags** - Losing kong's built-in validation | Keep existing validation logic in `config.Validate()` method; this is actually cleaner separation of concerns |

## Migration Plan

### Phase 1: Add Dependencies and Scaffolding (parallel with Phase 2)
1. Add `github.com/spf13/cobra` and `github.com/spf13/viper` to go.mod
2. Create new config loading function in `internal/config/load.go`
3. Maintain old kong-based logic temporarily

### Phase 2: Refactor Config Package
1. Update `Config` struct with `mapstructure` tags
2. Implement `loadFromViper()` function with file + env + defaults precedence
3. Wire in flag binding via cobra command
4. Keep validation logic intact (call existing `cfg.Validate()`)

### Phase 3: Update Main Entry Point
1. Create cobra root command in `cmd/landlord/main.go`
2. Define flags with cobra (--config, --database-host, etc.)
3. Add `PreRun` hook to load configuration via viper
4. Remove kong initialization code

### Phase 4: Testing and Validation
1. Write tests for viper loading with each precedence scenario
2. Test config file parsing (YAML and JSON)
3. Verify environment variables still work
4. Verify CLI flags still work and override correctly

### Phase 5: Deployment
1. Optional: Create example config.yaml and config.json
2. Update README with configuration guide
3. Deploy with old env vars (backward compatible)

**Rollback Strategy**: Kong struct remains importable until all references removed; can revert main.go and config.go if critical issues discovered.

## Open Questions

1. Should we auto-reload config files during runtime, or keep current "static at startup" behavior?
   - **Decision**: Static at startup for now; hot-reload is a future capability
   
2. What should happen if both config file and environment variables are provided?
   - **Decision**: Env vars override file (already specified in precedence)

3. Should we log the effective configuration at startup (with sources)?
   - **Decision**: Yes, for debugging; redact sensitive values (passwords, tokens)
