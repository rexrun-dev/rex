package generate

import "rexrun.dev/rex/internal/detect"

// CI generates a GitHub Actions CI workflow based on detected stack.
func CI(d *detect.Detection) string {
	switch d.Stack {
	case "go":
		return goCI()
	case "node":
		return nodeCI(d.PkgMgr)
	case "python":
		return pythonCI(d.PkgMgr)
	case "rust":
		return rustCI()
	case "java":
		return javaCI(d.PkgMgr)
	case "ruby":
		return rubyCI()
	case "php":
		return phpCI()
	case "elixir":
		return elixirCI()
	case "zig":
		return zigCI()
	default:
		return genericCI()
	}
}

func goCI() string {
	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - run: go test ./...
      - run: go vet ./...
      - run: go build ./...
`
}

func nodeCI(mgr string) string {
	install := "npm ci"
	test := "npm test"
	build := "npm run build --if-present"

	switch mgr {
	case "pnpm":
		install = "pnpm install --frozen-lockfile"
		test = "pnpm test"
		build = "pnpm run build --if-present"
	case "yarn":
		install = "yarn --frozen-lockfile"
		test = "yarn test"
		build = "yarn build --if-present"
	case "bun":
		install = "bun install --frozen-lockfile"
		test = "bun test"
		build = "bun run build"
	}

	setupMgr := ""
	if mgr == "pnpm" {
		setupMgr = `      - uses: pnpm/action-setup@v4
        with:
          version: latest
`
	}

	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
` + setupMgr + `      - uses: actions/setup-node@v4
        with:
          node-version: lts/*
      - run: ` + install + `
      - run: ` + test + `
      - run: ` + build + `
`
}

func pythonCI(mgr string) string {
	install := "pip install -r requirements.txt"
	test := "pytest"

	switch mgr {
	case "uv":
		install = "uv sync"
		test = "uv run pytest"
	case "poetry":
		install = "poetry install"
		test = "poetry run pytest"
	}

	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: "3.x"
      - run: ` + install + `
      - run: ` + test + `
`
}

func rustCI() string {
	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dtolnay/rust-toolchain@stable
      - run: cargo test
      - run: cargo clippy -- -D warnings
      - run: cargo fmt -- --check
`
}

func javaCI(mgr string) string {
	steps := "mvn test"
	if mgr == "gradle" {
		steps = "./gradlew test"
	}

	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-java@v4
        with:
          distribution: temurin
          java-version: 21
      - run: ` + steps + `
`
}

func rubyCI() string {
	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: ruby/setup-ruby@v1
        with:
          ruby-version: "3.3"
          bundler-cache: true
      - run: bundle exec rspec
`
}

func phpCI() string {
	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: shivammathur/setup-php@v2
        with:
          php-version: "8.3"
      - run: composer install --no-interaction
      - run: vendor/bin/phpunit
`
}

func elixirCI() string {
	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: erlef/setup-beam@v1
        with:
          otp-version: "26"
          elixir-version: "1.16"
      - run: mix deps.get
      - run: mix test
`
}

func zigCI() string {
	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: goto-bus-stop/setup-zig@v2
        with:
          version: master
      - run: zig build test
      - run: zig build
`
}

func genericCI() string {
	return `name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make test
`
}
