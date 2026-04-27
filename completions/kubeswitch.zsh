#compdef kubeswitch ks

# kubeswitch zsh completion

alias ks="kubeswitch"

_kubeswitch() {
    local -a contexts namespaces

    case ${CURRENT} in
        2)
            # First argument: suggest contexts
            contexts=(${(f)"$(kubectl config get-contexts -o name 2>/dev/null)"})
            _describe 'context' contexts
            ;;
        3)
            # Second argument: suggest namespaces for the given context
            namespaces=(${(f)"$(kubectl --context="${words[2]}" get namespaces -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null)"})
            _describe 'namespace' namespaces
            ;;
    esac
}

compdef _kubeswitch kubeswitch
compdef _kubeswitch ks
