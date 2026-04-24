# kubeswitch bash completion
# Source this file or add to your .bashrc / .bash_profile:
#   source <path-to>/completions/kubeswitch.bash

alias ks="kubeswitch"

_kubeswitch_autocomplete() {
    local cur prev contexts namespaces
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # First argument: suggest contexts
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        contexts=$(kubectl config get-contexts -o name 2>/dev/null)
        COMPREPLY=( $(compgen -W "${contexts}" -- "${cur}") )
        return 0
    fi

    # Second argument: suggest namespaces for the given context
    if [[ ${COMP_CWORD} -eq 2 ]]; then
        namespaces=$(kubectl --context="${prev}" get namespaces -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)
        COMPREPLY=( $(compgen -W "${namespaces}" -- "${cur}") )
        return 0
    fi
}

complete -F _kubeswitch_autocomplete kubeswitch
complete -F _kubeswitch_autocomplete ks
