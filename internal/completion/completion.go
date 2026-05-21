package completion

// Bash generates bash completion script.
func Bash() string {
	return `_rex() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="test run build deps clean fresh fmt lint clone init doctor help version"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands} --list --dry-run -v" -- "${cur}"))
        return 0
    fi

    case "${prev}" in
        clone)
            COMPREPLY=()
            ;;
        --dry-run)
            COMPREPLY=($(compgen -W "test run build deps clean fmt lint" -- "${cur}"))
            ;;
    esac
}
complete -F _rex rex
`
}

// Zsh generates zsh completion script.
func Zsh() string {
	return `#compdef rex

_rex() {
    local -a commands
    commands=(
        'test:run tests'
        'run:start the app'
        'build:build the project'
        'deps:install dependencies'
        'clean:remove build artifacts'
        'fresh:clean + deps + build'
        'fmt:format code'
        'lint:lint code'
        'clone:clone + detect + install deps'
        'init:generate rex.toml'
        'doctor:diagnose environment'
        '--list:show all detected commands'
        '--dry-run:show what would run'
        '-v:show version'
    )
    _describe 'command' commands
}

_rex "$@"
`
}

// Fish generates fish completion script.
func Fish() string {
	return `complete -c rex -f
complete -c rex -n "__fish_use_subcommand" -a "test" -d "Run tests"
complete -c rex -n "__fish_use_subcommand" -a "run" -d "Start the app"
complete -c rex -n "__fish_use_subcommand" -a "build" -d "Build the project"
complete -c rex -n "__fish_use_subcommand" -a "deps" -d "Install dependencies"
complete -c rex -n "__fish_use_subcommand" -a "clean" -d "Remove build artifacts"
complete -c rex -n "__fish_use_subcommand" -a "fresh" -d "Clean + deps + build"
complete -c rex -n "__fish_use_subcommand" -a "fmt" -d "Format code"
complete -c rex -n "__fish_use_subcommand" -a "lint" -d "Lint code"
complete -c rex -n "__fish_use_subcommand" -a "clone" -d "Clone + detect + install"
complete -c rex -n "__fish_use_subcommand" -a "init" -d "Generate rex.toml"
complete -c rex -n "__fish_use_subcommand" -a "doctor" -d "Diagnose environment"
complete -c rex -n "__fish_use_subcommand" -l "list" -d "Show all detected commands"
complete -c rex -n "__fish_use_subcommand" -l "dry-run" -d "Show what would run"
complete -c rex -n "__fish_use_subcommand" -s "v" -d "Show version"
`
}
