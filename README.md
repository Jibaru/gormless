# Gormless

[![Go Version](https://img.shields.io/badge/Go-1.24.5+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/Jibaru/gormless)](https://goreportcard.com/report/github.com/Jibaru/gormless)
[![Build Status](https://img.shields.io/github/actions/workflow/status/Jibaru/gormless/ci.yml?branch=main)](https://github.com/Jibaru/gormless/actions)
[![Release](https://img.shields.io/github/v/release/Jibaru/gormless?include_prereleases)](https://github.com/Jibaru/gormless/releases)
[![Docker Image](https://img.shields.io/badge/Docker-Available-2496ED?style=flat&logo=docker)](https://hub.docker.com/r/jibaru/gormless)

> A powerful, lightweight code generator that creates type-safe Data Access Objects (DAOs) for Go applications with support for PostgreSQL and MySQL.

## Features

- **Lightning Fast**: Generates comprehensive DAOs in seconds
- **Type-Safe**: Leverages Go's type system for compile-time safety
- **Multi-Database**: Supports PostgreSQL and MySQL with native query optimization
- **Zero Dependencies**: Generated code uses only standard library packages
- **Smart Tagging**: Automatic field mapping using struct tags
- **Transaction Support**: Built-in transaction management
- **Rich Operations**: CRUD, bulk operations, pagination, and counting
- **Clean Code**: Generates readable, maintainable Go code
- **CLI Ready**: Simple command-line interface

## Quick Start

### Installation

```bash
# Install via go install
go install github.com/Jibaru/gormless@latest

# Or download from releases
curl -L https://github.com/Jibaru/gormless/releases/latest/download/gormless-linux-amd64 -o gormless
chmod +x gormless
```

### Basic Usage

```bash
# Generate PostgreSQL DAOs
gormless -input ./models -output ./dao -driver postgres

# Generate MySQL DAOs  
gormless -input ./models/user.go -output ./dao -driver mysql
```

## Documentation

### Model Definition

Define your models with struct tags to specify database mapping:

```go
package models

import "time"

type User struct {
    ID        string    `sql:"id,primary"`
    Name      string    `sql:"name"`
    Email     string    `sql:"email"`
    Age       *int      `sql:"age"`
    CreatedAt time.Time `sql:"created_at"`
    UpdatedAt time.Time `sql:"updated_at"`
}

// Optional: Custom table name
func (u *User) TableName() string {
    return "users"
}
```

### Generated DAO

Gormless generates a comprehensive DAO with the following methods:

```go
type UserDAO struct {
    db *sql.DB
}

// CRUD Operations
func (dao *UserDAO) Create(ctx context.Context, user *User) error
func (dao *UserDAO) Update(ctx context.Context, user *User) error
func (dao *UserDAO) FindByPk(ctx context.Context, pk string) (*User, error)
func (dao *UserDAO) DeleteByPk(ctx context.Context, pk string) error

// Bulk Operations
func (dao *UserDAO) CreateMany(ctx context.Context, users []*User) error
func (dao *UserDAO) UpdateMany(ctx context.Context, users []*User) error
func (dao *UserDAO) DeleteManyByPks(ctx context.Context, pks []string) error

// Query Operations
func (dao *UserDAO) FindAll(ctx context.Context, where string, args ...interface{}) ([]*User, error)
func (dao *UserDAO) FindPaginated(ctx context.Context, limit, offset int, where string, args ...interface{}) ([]*User, error)
func (dao *UserDAO) Count(ctx context.Context, where string, args ...interface{}) (int64, error)

// Advanced Operations
func (dao *UserDAO) PartialUpdate(ctx context.Context, pk string, fields map[string]interface{}) error
func (dao *UserDAO) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
```

### Usage Example

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    "your-project/dao/postgres"
    "your-project/models"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "postgresql://user:password@localhost/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    userDAO := postgres.NewUserDAO(db)
    
    // Create a user
    user := &models.User{
        ID:    "user-123",
        Name:  "John Doe", 
        Email: "john@example.com",
    }
    
    err = userDAO.Create(context.Background(), user)
    if err != nil {
        log.Fatal(err)
    }
    
    // Find user by primary key
    foundUser, err := userDAO.FindByPk(context.Background(), "user-123")
    if err != nil {
        log.Fatal(err)
    }
    
    // Use transactions
    err = userDAO.WithTransaction(context.Background(), func(ctx context.Context) error {
        return userDAO.Update(ctx, foundUser)
    })
}
```

## Configuration

### Command Line Options

| Option | Short | Description | Required |
|--------|-------|-------------|----------|
| `--input` | `-i` | Path to model file or directory | ✅ |
| `--output` | `-o` | Output directory for generated DAOs | ✅ |
| `--driver` | `-d` | Database driver (`postgres`, `mysql`) | ✅ |

### Struct Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `sql:"column_name"` | Map field to database column | `sql:"user_name"` |
| `sql:"column_name,primary"` | Mark field as primary key | `sql:"id,primary"` |

### Database Support

| Database | Driver | Placeholder Style |
|----------|--------|------------------|
| PostgreSQL | `postgres` | `$1, $2, $3` |
| MySQL | `mysql` | `?, ?, ?` |

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the need for simple, efficient database access patterns in Go
- Built with [Cobra CLI](https://github.com/spf13/cobra) for excellent command-line experience

## Links

- [Documentation](https://github.com/Jibaru/gormless/wiki)
- [Examples](https://github.com/Jibaru/gormless/tree/main/examples)
- [Issue Tracker](https://github.com/Jibaru/gormless/issues)
- [Releases](https://github.com/Jibaru/gormless/releases)
