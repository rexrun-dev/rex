package generate

import (
	"fmt"
	"strings"

	"rexrun.dev/rex/internal/detect"
)

// Dockerfile generates a Dockerfile for the detected stack.
func Dockerfile(d *detect.Detection) string {
	switch d.Stack {
	case "go":
		return goDockerfile()
	case "node":
		return nodeDockerfile(d.PkgMgr)
	case "python":
		return pythonDockerfile(d.PkgMgr)
	case "rust":
		return rustDockerfile()
	case "java":
		return javaDockerfile(d.PkgMgr)
	case "ruby":
		return rubyDockerfile()
	case "php":
		return phpDockerfile()
	case "elixir":
		return elixirDockerfile()
	default:
		return genericDockerfile()
	}
}

func goDockerfile() string {
	return `# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./...

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["./app"]
`
}

func nodeDockerfile(mgr string) string {
	installCmd := "npm ci"
	buildCmd := "npm run build"
	startCmd := "npm start"

	switch mgr {
	case "pnpm":
		installCmd = "pnpm install --frozen-lockfile"
		buildCmd = "pnpm run build"
		startCmd = "pnpm start"
	case "yarn":
		installCmd = "yarn install --frozen-lockfile"
		buildCmd = "yarn build"
		startCmd = "yarn start"
	case "bun":
		installCmd = "bun install"
		buildCmd = "bun run build"
		startCmd = "bun start"
	}

	return `# Build stage
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN ` + installCmd + `
COPY . .
RUN ` + buildCmd + `

# Runtime stage
FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY package*.json ./
RUN ` + installCmd + `
CMD ["` + strings.TrimPrefix(startCmd, "npm ") + `"]
`
}

func pythonDockerfile(mgr string) string {
	install := "pip install -r requirements.txt"
	cmd := "python main.py"

	switch mgr {
	case "uv":
		install = "pip install uv && uv sync"
		cmd = "uv run python main.py"
	case "poetry":
		install = "pip install poetry && poetry install --no-dev"
		cmd = "poetry run python main.py"
	}

	return `FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt ./
RUN ` + install + `
COPY . .
CMD ["` + cmd + `"]
`
}

func rustDockerfile() string {
	return `# Build stage
FROM rust:1.85-slim AS builder
WORKDIR /app
COPY Cargo.toml Cargo.lock ./
COPY src ./src
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/app /usr/local/bin/app
CMD ["app"]
`
}

func javaDockerfile(mgr string) string {
	build := "mvn package -DskipTests"
	jar := "target/*.jar"
	if mgr == "gradle" {
		build = "./gradlew build -x test"
		jar = "build/libs/*.jar"
	}

	return `# Build stage
FROM eclipse-temurin:21-jdk-alpine AS builder
WORKDIR /app
COPY . .
RUN ` + build + `

# Runtime stage
FROM eclipse-temurin:21-jre-alpine
WORKDIR /app
COPY --from=builder /app/` + jar + ` app.jar
CMD ["java", "-jar", "app.jar"]
`
}

func rubyDockerfile() string {
	return `FROM ruby:3.3-slim
WORKDIR /app
COPY Gemfile Gemfile.lock ./
RUN bundle install
COPY . .
CMD ["bundle", "exec", "rails", "server", "-b", "0.0.0.0"]
`
}

func phpDockerfile() string {
	return `FROM php:8.3-apache
WORKDIR /var/www/html
COPY composer.json composer.lock ./
RUN apt-get update && apt-get install -y unzip && docker-php-ext-install pdo pdo_mysql
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer
RUN composer install --no-interaction
COPY . .
`
}

func elixirDockerfile() string {
	return `# Build stage
FROM elixir:1.17-alpine AS builder
WORKDIR /app
RUN mix local.hex --force && mix local.rebar --force
COPY mix.exs mix.lock ./
RUN mix deps.get --only prod
COPY . .
RUN mix compile

# Runtime stage
FROM elixir:1.17-alpine
WORKDIR /app
COPY --from=builder /app/_build /app/_build
COPY --from=builder /app/deps /app/deps
CMD ["mix", "phx.server"]
`
}

func genericDockerfile() string {
	return `# Detected stack: unknown
# Customize this Dockerfile for your project
FROM ubuntu:24.04
WORKDIR /app
COPY . .
RUN make deps && make build
CMD ["make", "run"]
`
}

// Badge generates a markdown badge for the detected stack.
func Badge(d *detect.Detection) string {
	label := d.Stack
	if d.PkgMgr != "" && d.PkgMgr != d.Stack {
		label += " + " + d.PkgMgr
	}
	return fmt.Sprintf(`[![rex](https://img.shields.io/badge/rex-%s-orange)](https://github.com/rexrun-dev/rex)`, label)
}
