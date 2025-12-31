# Oh Man! Bash Completion Script
# Installation: source this file, or add the contents to ~/.bashrc

_ohman_completions() {
    local cur prev opts commands subcommands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Main options
    opts="-s --section -m --model -r --raw -i --interactive -c --config -v --verbose -h --help --version"
    
    # Subcommands
    subcommands="config history clear"

    # Handle option arguments
    case $prev in
        -s|--section)
            COMPREPLY=( $(compgen -W "1 2 3 4 5 6 7 8" -- "$cur") )
            return 0
            ;;
        -m|--model)
            COMPREPLY=( $(compgen -W "gpt-4o gpt-4o-mini gpt-4 claude-3-sonnet claude-3-haiku llama3 mistral" -- "$cur") )
            return 0
            ;;
        -c|--config)
            COMPREPLY=( $(compgen -f -- "$cur") )
            return 0
            ;;
    esac

    # If current word starts with -, complete options
    if [[ $cur == -* ]]; then
        COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
        return 0
    fi

    # If first non-option argument, complete subcommands or system commands
    local arg_count=0
    for word in "${COMP_WORDS[@]:1}"; do
        if [[ $word != -* ]] && [[ $word != "${COMP_WORDS[COMP_CWORD]}" ]]; then
            ((arg_count++))
        fi
    done

    if [[ $arg_count -eq 0 ]]; then
        # Subcommands
        local sub_completions=$(compgen -W "$subcommands" -- "$cur")
        # System commands (limited count)
        local cmd_completions=$(compgen -c "$cur" 2>/dev/null | head -20)
        COMPREPLY=( $sub_completions $cmd_completions )
        return 0
    fi
}

complete -F _ohman_completions ohman
