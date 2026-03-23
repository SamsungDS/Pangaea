#!/bin/bash
# Show pods info from <worker node>

function usage() {
    echo "<Usage>
        $0 <Worker Node>
    "
    exit 1
}

#if [ "$#" != 1 ]; then
#  usage
#fi

#NODE=$1
NODE=gigabyte-turin

readonly SCRIPT_PATH=$(dirname $(readlink -f "${BASH_SOURCE[0]}"))

watch -t -n 0.1 -d $SCRIPT_PATH/show_available_memory.sh $NODE
