# ESLint Custom Rule Pattern

## When to Use
Use this pattern for creating custom ESLint rules that enforce project-specific coding standards, prevent common mistakes, and ensure consistent component composition patterns.

## Why It Exists
This pattern provides automated enforcement of architectural decisions and best practices that cannot be caught by standard linting rules, particularly around component composition and accessibility.

## Implementation
Custom ESLint rules are implemented as JavaScript modules with specific structure. Key characteristics:

- Rule definition with meta information including type, docs, and messages
- Context-based AST traversal for pattern detection
- Component stack tracking for nested element analysis
- Comprehensive component type lists for different categories
- Clear error messages with actionable fix suggestions
- Support for both simple and dot notation component names

## References
- `frontend/eslint-plugin-local/index.js:3-16` - Rule metadata with documentation and error messages
- `frontend/eslint-plugin-local/index.js:22-39` - Trigger components list for interactive element detection
- `frontend/eslint-plugin-local/index.js:42-51` - Interactive components that should not be nested
- `frontend/eslint-plugin-local/index.js:53-65` - Component name extraction for different syntax patterns
- `frontend/eslint-plugin-local/index.js:67-108` - AST traversal logic with stack-based detection
- `frontend/eslint.config.js` - Integration of custom plugin into ESLint configuration

## Key Conventions
- Use descriptive rule names that clearly indicate the prohibited pattern
- Provide comprehensive documentation with examples of before/after fixes
- Include both standalone and dot notation component names in detection lists
- Use clear, actionable error messages that guide developers to correct solutions
- Implement proper stack management for nested component detection
- Test rules against real component patterns in the codebase
- Group related component types logically (triggers, interactive elements, etc.)