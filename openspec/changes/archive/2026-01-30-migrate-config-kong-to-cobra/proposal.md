## Why

The current configuration system uses alexthomas/kong for CLI flag parsing and environment variable loading. While kong provides basic functionality, adopting cobra/viper would provide more flexibility, better ecosystem integration, and superior support for complex configuration scenarios. Cobra is the de facto standard for CLI applications in Go (used by kubectl, Docker, Hugo), while viper enables seamless multi-format configuration support (YAML, JSON, TOML, etc.) with environment variable and CLI flag overrides following standard precedence rules.

## What Changes

- **Replace kong CLI parser**: Migrate from `alexthomas/kong` to `cobra/cobra` for command and flag parsing
- **Add viper configuration**: Integrate `spf13/viper` for unified configuration management with support for multiple file formats
- **Support YAML and JSON config files**: Enable configuration via `config.yaml` or `config.json` files in addition to environment variables and CLI flags
- **Establish configuration precedence**: Implement clear precedence order: CLI flags > environment variables > config file > defaults
- **Update configuration loading**: Refactor the config loading logic in `internal/config/` to use cobra/viper instead of kong
- **Maintain backward compatibility**: Preserve environment variable names and meanings so existing deployments continue working
- **Update documentation**: Document new configuration file formats, precedence rules, and migration path

## Capabilities

### New Capabilities
- `cobra-cli-framework`: Support for cobra command structure with proper help text, subcommands, and command discovery
- `viper-configuration-files`: Multi-format configuration file support (YAML, JSON) with hot-reload awareness and structured binding

### Modified Capabilities
- `configuration-management`: Change from kong-based to cobra/viper-based configuration loading while maintaining environment variable support and CLI flag overrides

## Impact

- **Code**: `internal/config/config.go` and `cmd/landlord/main.go` require refactoring
- **Dependencies**: Remove `github.com/alecthomas/kong`, add `github.com/spf13/cobra` and `github.com/spf13/viper`
- **User Impact**: Users can now use YAML/JSON config files; CLI flags and environment variables remain unchanged
- **Deployment**: No breaking changes to environment variables; config file support is optional
- **Testing**: Updated configuration tests to cover cobra/viper patterns and file-based configuration
