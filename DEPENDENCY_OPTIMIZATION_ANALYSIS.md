# Dependency Optimization Analysis

## Current State
- Binary size: 9.9MB (not 14MB as initially thought)
- Main external dependencies in cmd:
  - koanf (with multiple sub-packages) - configuration management
  - kong - CLI parsing
  - kongplete - shell completion
  - yaml.v3 - YAML parsing (used by koanf)

## Optimization Options

### Option 1: Remove Koanf, Keep Kong (Moderate Optimization)
**Changes:**
- Replace koanf with simple YAML unmarshaling + manual env loading
- Keep kong for structured CLI parsing
- Keep kongplete for shell completion

**Dependencies saved:**
- koanf/v2 and all subpackages
- mapstructure (indirect)
- mitchellh/copystructure (indirect)
- fatih/structs (indirect)

**Estimated size reduction:** ~1-2MB

**Implementation:** See `config_simple.go`

### Option 2: Remove Kongplete (Minor Optimization)
**Changes:**
- Remove automatic shell completion
- Implement basic manual completion or remove feature
- Keep kong and koanf

**Dependencies saved:**
- kongplete
- posener/complete
- riywo/loginshell

**Estimated size reduction:** ~0.5MB

### Option 3: Full Minimal (Major Optimization)
**Changes:**
- Remove koanf - use simple YAML + env
- Remove kongplete - no shell completion
- Keep only kong for CLI parsing
- Alternative: Use flag package from stdlib

**Dependencies saved:**
- All koanf packages
- kongplete and deps
- Many indirect dependencies

**Estimated size reduction:** ~2-3MB

### Option 4: Ultra Minimal (Maximum Optimization)
**Changes:**
- Replace kong with stdlib flag package
- Simple YAML config with yaml.v3
- Manual env var handling
- No shell completion

**Dependencies:**
- Only yaml.v3 for config files
- Everything else from stdlib

**Estimated size reduction:** ~3-4MB

## Recommendations

1. **Best Balance (Option 1):** Remove koanf, keep kong/kongplete
   - Maintains good CLI UX with completion
   - Reduces complexity in config management
   - Moderate size reduction

2. **If size is critical (Option 4):** Full stdlib approach
   - Minimal dependencies
   - Maximum size reduction
   - Loss of some CLI conveniences

3. **Current approach is reasonable** if:
   - 9.9MB binary size is acceptable
   - You value the features koanf provides
   - Shell completion is important

## Trade-offs

### Removing Koanf
**Pros:**
- Reduces dependencies significantly
- Simpler code for basic config needs
- Removes mapstructure dependency

**Cons:**
- Lose flexible configuration providers
- Need to implement env var mapping manually
- Less extensible for future config sources

### Removing Kongplete
**Pros:**
- Fewer dependencies
- Simpler build

**Cons:**
- No automatic shell completion
- Worse CLI UX for power users

### Removing Kong
**Pros:**
- Significant dependency reduction
- Smaller binary

**Cons:**
- Lose type-safe CLI parsing
- Need to implement command routing manually
- Less self-documenting CLI structure