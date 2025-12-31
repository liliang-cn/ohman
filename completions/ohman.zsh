# Oh Man! Zsh Completion Script
# Installation: place this file in $fpath directory, or source it

#compdef ohman

_ohman() {
    local curcontext="$curcontext" state line
    typeset -A opt_args

    _arguments -C \
        '-s+[Specify man page section (1-8)]:section:(1 2 3 4 5 6 7 8)' \
        '--section+[Specify man page section (1-8)]:section:(1 2 3 4 5 6 7 8)' \
        '-m+[Specify LLM model]:model:' \
        '--model+[Specify LLM model]:model:' \
        '-r[Show raw man content without AI]' \
        '--raw[Show raw man content without AI]' \
        '-i[Force interactive mode]' \
        '--interactive[Force interactive mode]' \
        '-c+[Config file path]:config:_files' \
        '--config+[Config file path]:config:_files' \
        '-v[Verbose output mode]' \
        '--verbose[Verbose output mode]' \
        '-h[Show help]' \
        '--help[Show help]' \
        '--version[Show version]' \
        '1: :->command' \
        '*: :->args'

    case $state in
        command)
            # Provide subcommands or system commands as completions
            local -a subcommands
            subcommands=(
                'config:Configure ohman'
                'history:View session history'
                'clear:Clear session cache'
            )
            
            # If input doesn't start with -, provide man-available commands
            if [[ $words[CURRENT] != -* ]]; then
                # Get common commands as completions
                local -a commands
                commands=($(compgen -c 2>/dev/null | head -100))
                _describe -t commands 'commands' commands
                _describe -t subcommands 'subcommands' subcommands
            fi
            ;;
        args)
            # If first argument is a subcommand, don't provide further completions
            case $words[2] in
                config|history|clear)
                    ;;
                *)
                    # Provide question example
                    _message "Enter your question (e.g., How to use the -r flag?)"
                    ;;
            esac
            ;;
    esac
}

_ohman "$@"
