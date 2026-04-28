# GoProfiler - Learning Resources

A curated list of resources to understand the concepts used in this project.

---

## Table of Contents

1. [Go Fundamentals](#go-fundamentals)
2. [Go Code Analysis](#go-code-analysis)
3. [CLI Development](#cli-development)
4. [Testing](#testing)
5. [CI/CD & Releases](#cicd--releases)
6. [Project Structure](#project-structure)

---

## Go Fundamentals

### Go Modules
- **Official Doc**: https://go.dev/blog/using-go-modules
- **Video**: https://www.youtube.com/watch?v=MSdist2zr7Iw
- **Blog**: https://ieft.in/articles/golang-modules-explained/

### Interfaces
- **Official Tour**: https://go.dev/tour/methods/9
- **Video**: https://www.youtube.com/watch?v=yl4ACwtM0gM
- **Blog**: https://gobyexample.com/interfaces

### init() Function
- **Official Doc**: https://go.dev/ref/spec#Package_initialization
- **Blog**: https://medium.com/golang-examples/init-functions-in-go-4fd9f08e4a0

### Pointers & Memory
- **Official Tour**: https://go.dev/tour/moretypes/1
- **Video**: https://www.youtube.com/watch?v=9jJ3xY4q4gI
- **Blog**: https://gobyexample.com/pointers

---

## Go Code Analysis

### AST (Abstract Syntax Tree)
- **Official Package**: https://pkg.go.dev/go/ast
- **Video**: https://www.youtube.com/watch?v=LOnn2TId0Bk
- **Blog**: https://dominikbraun.io/blog/understanding-go-ast

### go/ast Package
- **Official Docs**: https://pkg.go.dev/go/ast
- **Blog**: https://blog.gopheracademy.com/advent-2014/ast-rewriting/
- **Examples**: https://github.com/golang/example/tree/master/gotypes

### go/parser Package
- **Official Docs**: https://pkg.go.dev/go/parser
- **Blog**: https://pkg.go.dev/go/parser#ParseFile

### go/token Package
- **Official Docs**: https://pkg.go.dev/go/token
- **Blog**: https://ieft.in/articles/go-token-package/

### Regex Patterns
- **Official Doc**: https://pkg.go.dev/regexp
- **Video**: https://www.youtube.com/watch?v=OSj1bZ9NQz0
- **Cheatsheet**: https://github.com/cedrickchee/awesome-go#regular-expression

---

## CLI Development

### urfave/cli
- **Official Repo**: https://github.com/urfave/cli
- **Docs**: https://cli.urfave.org/
- **Examples**: https://github.com/urfave/cli/tree/main/examples

### Building CLI Tools in Go
- **Blog Series**: https://blog.alexellis.io/golang-cli-tool/
- **Video**: https://www.youtube.com/watch?v=aC7pP8k2tT0
- **Book**: https://www.manning.com/books/command-line-applications-in-go

---

## Testing

### Go Testing Package
- **Official Doc**: https://pkg.go.dev/testing
- **Video**: https://www.youtube.com/watch?v=ysFPgRHkAnc
- **Blog**: https://gobyexample.com/testing

### Table-Driven Tests
- **Blog**: https://dave.cheney.net/2019/10/08/visualising-table-driven-tests
- **Examples**: https://github.com/golang/go/wiki/TableDrivenTests

### Mocking
- **Blog**: https://medium.com/@rochakjain361/mocking-in-go-4d2735b84f70
- **Library**: https://github.com/golang/mock

---

## CI/CD & Releases

### GitHub Actions
- **Official Docs**: https://docs.github.com/en/actions
- **Video**: https://www.youtube.com/watch?v=0squX_5aVh4
- **Blog**: https://inercept.in/blogs/github-actions-for-go/

### GoReleaser
- **Official Docs**: https://goreleaser.com/
- **Video**: https://www.youtube.com/watch?v=flZ0Xz0t2Ds
- **Blog**: https://blog.URENTAL.dev/goreleaser-getting-started

### GolangCI-Lint
- **Official Repo**: https://github.com/golangci/golangci-lint
- **Docs**: https://golangci-lint.run/

---

## Project Structure

### Standard Go Project Layout
- **Blog**: https://github.com/golang-standards/project-layout
- **Article**: https://medium.com/golang-projects/clean-architecture-in-go-af09c48c8c4f

### cmd/ vs internal/
- **Issue Discussion**: https://github.com/golang/go/issues/27359
- **Blog**: https://ieft.in/articles/go-internal-packages/

---

## Architecture Patterns

### Plugin/Registry Pattern
- **Video**: https://www.youtube.com/watch?v=NGM2qICjYJQ
- **Blog**: https://dev.to/koddr/building-pluggable-golang-applications-1l8b
- **Example**: https://github.com/hashicorp/go-plugin

### Strategy Pattern
- **Blog**: https://refactoring.guru/design-patterns/strategy/go/example
- **Video**: https://www.youtube.com/watch?v=v9xO7MHNN5M

---

## Practice Exercises

### Beginner
1. Parse a Go file and print all function names
2. Create a simple CLI with urfave/cli
3. Write a table-driven test

### Intermediate
1. Write an AST pattern to detect empty loops
2. Add a new pattern to this project
3. Create a custom output formatter

### Advanced
1. Implement code transformation (not just detection)
2. Add auto-fix generation
3. Build a plugin system with go-plugin

---

## Recommended Learning Path

| Day | Topic | Time |
|-----|-------|------|
| 1 | Go basics + modules | 2 hrs |
| 2 | Interfaces + init() | 1 hr |
| 3 | go/ast + go/parser | 3 hrs |
| 4 | Build a simple analyzer | 2 hrs |
| 5 | CLI with urfave/cli | 2 hrs |
| 6 | Testing + CI/CD | 2 hrs |

---

## Quick Reference

### Key Packages
```bash
go get golang.org/x/lint/golint    # Linting
go get github.com/golang/mock/mockgen  # Mocking
go get github.com/golangci/golangci-lint  # CI linting
go get github.com/goreleaser/goreleaser  # Releases
go get github.com/urfave/cli/v2       # CLI framework
```

### Useful Commands
```bash
go mod init <module-name>
go mod tidy
go build ./...
go test ./...
go test -v -race ./...
goreleaser init
```

---

## Community

- **Gophercises**: https://gophercises.com/ (Free exercises)
- **Go by Example**: https://gobyexample.com/
- **Go Wiki**: https://github.com/golang/go/wiki/Learn
- **Golang Weekly**: https://golangweekly.com/
- **r/golang**: https://reddit.com/r/golang
