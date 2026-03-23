#!/bin/bash

function usage() {
    echo "<Usage>
        $0 <Worker Node IP>
    "
    exit 1
}

if [ "$#" != 1 ]; then
	usage
fi

lsmem_cmd='watch -t -n 0.1 -d "echo \# lsmem --output-all;echo; lsmem --output-all"'

ssh -t root@$1 $lsmem_cmd 2>/dev/null