#!/bin/bash
# Show pods info from <worker node>

function usage() {
    echo "<Usage>
        $0 <Worker Node>
    "
    exit 1
}

#if [ "$#" != 1 ]; then
#	usage
#fi

#NODE=$1
NODE=gigabyte-turin

watch -t -n 0.1 -d "echo \# kubectl get pod -o wide;echo; kubectl get pod --field-selector spec.nodeName=$NODE -o wide"
