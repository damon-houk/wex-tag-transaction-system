# Contributing to WEX TAG Transaction Processing System

Thank you for considering contributing to this project! This document outlines the branching strategy, workflow, and guidelines for contributing to the repository.

## Branching Strategy

We follow a simplified version of GitHub Flow with the following branch structure:

### Main Branch (Protected)
- Always represents production-ready code
- Direct commits are blocked; all changes come through pull requests
- Should be deployable at any time

### Feature Branches
- Created for each new feature or bug fix
- Named descriptively, e.g., `feature/currency-converter` or `fix/validation-error`
- Short-lived and frequently merged back to main

### Release Branches (Optional)
- Created when preparing for a specific release
- Named with version numbers, e.g., `release/1.0.0`
- Used for final testing and bug fixes before deployment

## Development Workflow

1. **Create a branch**:# Contributing to WEX TAG Transaction Processing System

Thank you for considering contributing to this project! This document outlines the branching strategy, workflow, and guidelines for contributing to the repository.

## Branching Strategy

We follow a simplified version of GitHub Flow with the following branch structure:

### Main Branch (Protected)
- Always represents production-ready code
- Direct commits are blocked; all changes come through pull requests
- Should be deployable at any time

### Feature Branches
- Created for each new feature or bug fix
- Named descriptively, e.g., `feature/currency-converter` or `fix/validation-error`
- Short-lived and frequently merged back to main

### Release Branches (Optional)
- Created when preparing for a specific release
- Named with version numbers, e.g., `release/1.0.0`
- Used for final testing and bug fixes before deployment

## Development Workflow

1. **Create a branch**:
   ```
   git checkout -b feature/your-feature-name
   ```

2. **Make changes**: Implement your feature or fix the bug.

3. **Write tests**: Ensure your code is well-tested.

4. **Run tests locally**:
   ```
   make test
   ```

5. **Commit your changes**:
   ```
   git add .
   git commit -m "Description of changes"
   ```

6. **Push to your branch**:
   ```
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request**: Submit a PR from your branch to `main`.

8. **Code Review**: Wait for code review and address any feedback.

9. **Merge**: Once approved, the changes will be merged to `main`.

## Pull Request Guidelines

When submitting a pull request, please ensure:

1. The code follows Go best practices and the project's style guide
2. All tests pass
3. New code has appropriate test coverage
4. Documentation is updated if necessary
5. The PR description clearly explains the changes

## Code Style

Follow the standard Go code style and formatting guidelines. Run `go fmt` before committing.

## Testing

Write tests for all new code. We aim for a high level of test coverage:

- Unit tests for business logic
- Integration tests for API endpoints
- Performance tests for critical paths

## Commit Messages

Write clear, concise commit messages that explain the "what" and "why" of the changes:

```
Short summary (72 chars or less)

More detailed explanation if necessary. Wrap lines at 72 characters.
Explain the problem that this commit is solving and how it solves it.

Closes #123
```

Thank you for helping improve the WEX TAG Transaction Processing System!