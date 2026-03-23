#!/bin/bash
# Apply <pod> to <worker node>

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
pod_yaml="$1-60gb-$2.yaml"

if [ -e "$YAML_PATH/$pod_yaml" ]; then
	if [ "$3" != "silent" ]; then
		printf "kubectl apply -f %s\n" "$YAML_PATH/$pod_yaml"
		kubectl apply -f "$YAML_PATH/$pod_yaml"
	else
		kubectl apply -f "$YAML_PATH/$pod_yaml" &> /dev/null
	fi
else
	echo "Pod Spec yaml does not exist: $YAML_PATH/$pod_yaml"
	usage
fi
