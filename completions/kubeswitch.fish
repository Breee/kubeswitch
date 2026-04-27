# kubeswitch fish completion

alias ks="kubeswitch"

# First argument: contexts
complete -c kubeswitch -f -n "__fish_is_first_arg" -a "(kubectl config get-contexts -o name 2>/dev/null)"
complete -c ks -f -n "__fish_is_first_arg" -a "(kubectl config get-contexts -o name 2>/dev/null)"

# Second argument: namespaces for the given context
complete -c kubeswitch -f -n "not __fish_is_first_arg" -a "(kubectl --context=(commandline -opc)[2] get namespaces -o jsonpath='{range .items[*]}{.metadata.name}\n{end}' 2>/dev/null)"
complete -c ks -f -n "not __fish_is_first_arg" -a "(kubectl --context=(commandline -opc)[2] get namespaces -o jsonpath='{range .items[*]}{.metadata.name}\n{end}' 2>/dev/null)"

function __fish_is_first_arg
    set -l tokens (commandline -opc)
    test (count $tokens) -eq 1
end
