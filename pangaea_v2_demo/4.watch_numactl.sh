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

numactl_cmd='watch -t -n 0.1 -d "echo \# numactl --hardware;echo; numactl --hardware"'

ssh -t root@$1 $numactl_cmd 2>/dev/null