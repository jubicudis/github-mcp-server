# 7D Context Framework Implementation

## Overview

This document describes the implementation of the 7D Context Framework in the Tranquility Neuro-OS (TNOS) project. The framework provides a standardized approach to contextual awareness across all system components.

## Core Components

The 7D Context Framework consists of these dimensions:

1. **WHO**: Actor & Identity Context
2. **WHAT**: Function & Content Context
3. **WHEN**: Temporal Context
4. **WHERE**: Location Context
5. **WHY**: Intent & Purpose Context
6. **HOW**: Method & Process Context
7. **TO WHAT EXTENT**: Scope & Impact Context

## File Structure

The implementation is organized across the following files:

- **`context.go`**: Main implementation of the 7D Context Framework
- **`context_bridge.go`**: Bridge for compatibility between different context implementations
- **`translations.go`**: Additional translation utilities (with legacy context implementation)
- **`common.go`**: Shared utilities for translations

## Main Implementation

The primary implementation in `context.go` provides:

- `ContextVector7D` structure with all 7 dimensions
- Compression and decompression using the Möbius Compression Formula
- Context translation between GitHub MCP and TNOS formats
- Helper functions for parameter extraction, validation, etc.

## Context Bridge

The `context_bridge.go` file provides compatibility between different context implementations:

- Bidirectional synchronization between implementations
- Migration utilities for standardizing on the `context.go` implementation
- Backward compatibility functions for legacy code

## Usage Guidelines

### New Code

All new code should use the `context.go` implementation:

```go
// Create a new context
cv := translations.NewContext("Actor", "Operation", "Location", "Purpose", "Method", 1.0)

// Store in Go context
ctx = translations.ContextWithVector(ctx, cv)

// Retrieve from Go context
cv, exists := translations.VectorFromContext(ctx)
```

### Legacy Support

For code that might use the older implementation, use the bridge functions:

```go
// Synchronize context between implementations
ctx = translations.BridgeContextsAndSync(ctx)

// Migrate all references to use the standardized implementation
ctx = translations.MigrateAllContextReferences(ctx, logger)
```

## Möbius Compression Formula

The framework implements compression-first logic using the Möbius Compression Formula:

```math
compressed = (value * B * I * (1 - (entropy / log2(1 + V))) * (G + F))
/ (E * t + C_sum * entropy + alignment)

alignment = (B + V * I) * exp(-t * E)
```

Where:

- B: Base factor (0.8)
- V: Value factor (0.7)
- I: Intent factor (0.9)
- G: Growth factor (1.2)
- F: Flexibility factor (0.6)
- E: Energy factor (0.5)
- t: Time factor (calculated based on age)

## Best Practices

1. Always use the 7D Context Framework in function signatures
2. Document code with the 7D framework comments
3. Apply compression-first logic using the provided methods
4. Include all necessary context dimensions
5. Use the bridge functions when interfacing with legacy code
6. Ensure proper context transmission across system boundaries

## Future Directions

1. Complete migration to the standardized implementation
2. Enhance compression algorithms based on real-world usage patterns
3. Implement context-aware caching for performance optimization
4. Extend the bridge to support additional context systems
5. Add monitoring and diagnostics for context operations
