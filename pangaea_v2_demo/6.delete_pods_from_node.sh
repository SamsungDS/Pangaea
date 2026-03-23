#!/bin/bash
# more pods on CXL Host than Normal Host

readonly SCRIPT_PATH=$(dirname $(readlink -f "${BASH_SOURCE[0]}"))
readonly DELETE_SCRIPT="$SCRIPT_PATH/delete_pod.sh"

function pause(){
	read -n 1 -s -r -p "Press any key to continue..."
	echo
}

function delete_pod(){
	$DELETE_SCRIPT $1 $2
}

NODE=gigabyte-turin

printf "[Delete all pods from $NODE]\n"
for ((i=1;i<=5;i++))
do
	delete_pod $NODE $i
done
