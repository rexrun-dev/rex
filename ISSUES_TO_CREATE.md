# GitHub Issues da creare (per attirare contributor)

## Good First Issues

### 1. Add PHP/Laravel detection
**Labels**: good first issue, enhancement
**Body**: 
Rex should detect PHP projects with `composer.json` and Laravel projects with `artisan`.
- `rex test` → `php artisan test` or `./vendor/bin/phpunit`
- `rex run` → `php artisan serve`
- `rex deps` → `composer install`

### 2. Add Ruby/Rails detection
**Labels**: good first issue, enhancement
**Body**:
Rex should detect Ruby projects with `Gemfile` and Rails with `config/routes.rb`.
- `rex test` → `bundle exec rspec` or `rails test`
- `rex run` → `rails server`
- `rex deps` → `bundle install`

### 3. Add Zig detection
**Labels**: good first issue, enhancement
**Body**:
Rex should detect Zig projects with `build.zig`.
- `rex test` → `zig build test`
- `rex build` → `zig build`
- `rex run` → `zig build run`

### 4. Add Elixir/Phoenix detection
**Labels**: good first issue, enhancement
**Body**:
Rex should detect Elixir projects with `mix.exs`.
- `rex test` → `mix test`
- `rex run` → `mix phx.server`
- `rex deps` → `mix deps.get`

### 5. Add shell completions (bash/zsh/fish)
**Labels**: good first issue, enhancement
**Body**:
Add `rex --completions bash|zsh|fish` to generate shell completion scripts.
Commands to complete: test, run, build, deps, clean, fresh, fmt, lint, doctor

### 6. Add --dry-run flag
**Labels**: good first issue, enhancement
**Body**:
Add a `--dry-run` flag that shows what command would be executed without actually running it.
`rex test --dry-run` → prints `would run: go test ./...`

## Feature Requests

### 7. Add monorepo support
**Labels**: enhancement
**Body**:
In monorepos, rex should detect the current subdirectory's stack.
If you're in `packages/api/` with its own `package.json`, rex should use that.

### 8. Add .env file loading
**Labels**: enhancement  
**Body**:
Rex could automatically load `.env` file before running commands.
`rex run` in a Node project could auto-load `.env` with dotenv behavior.
