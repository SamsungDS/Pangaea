#!/bin/bash
# Delete <pod> from <worker node>

function usage() {
    echo "<Usage>
        $0 <Worker Node> <POD_NUM>
    "
    exit 1
}

if [ $# -lt 2 ]; then
	usage
fi

readonly SCRIPT_PATH=$(dirname $(readlink -f "${BASH_SOURCE[0]}"))
readonly YAML_PATH="$SCRIPT_PATH/yaml/$1"
pod_name="$1-60gb-$2"

is_pod_exist=$(kubectl get pod -A | grep $pod_name)

force=""
if [ "$4" == "force" ]; then
	force="--force"
fi

if [ "x$is_pod_exist" != "x" ]; then
	if [ "$3" != "silent" ]; then
		printf "kubectl delete pod %s\n" "$pod_name"
		kubectl delete -f "$YAML_PATH/$pod_name.yaml" $force
	else
		kubectl delete -f "$YAML_PATH/$pod_name.yaml" $force &> /dev/null
	fi
else
	printf "Pod does not exist: %s\n" "$pod_name"
	usage
fi
