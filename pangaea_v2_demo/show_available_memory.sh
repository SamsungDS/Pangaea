#!/bin/bash

# ./show_available_memory.sh <NodeName> | all
# display node's available memory

usage() {
    echo "<Usage>
        $0 <NodeName>
	or
	$0 all
    "
}

function to_bytes() {
	## $1: initial value / $2: initial value's unit / $3: converted value as bytes
	local init_value=$1
	local -n unit=$2
	local -n bytes_value=$3
	local units=("Ki" "Mi" "Gi" "Ti" "Pi" "Ei" "Zi" "Yi")
	local i=0

	local init_value_mem=$(echo $init_value | grep -o '[0-9]' | tr -d '\n')

	if [[ $init_value =~ ^[0-9]+$ ]]; then
		unit="B"
		bytes_value=$init_value_mem
		return 0
	else
		init_value_mem=$(( init_value_mem * 1024 ))
	fi

	while [[ $init_value != *"${units[$i]}" ]] && (( i < ${#units[@]} - 1 ))
	do
		init_value_mem=$(( init_value_mem * 1024 ))
		((i++))
	done

	unit=${units[$i]}
	bytes_value=$init_value_mem
}

function align_unit() {
	## $1: bytes value / $2: initial value's unit / $3: converted value as bytes
	local bytes_value=$1
	local unit_align=$2
	local -n value_align=$3
	local units=("B" "KiB" "MiB" "GiB" "TiB" "PiB" "EiB" "ZiB" "YiB")
	local i=0

	while [[ $unit_align != ${units[$i]} ]] && (( i < ${#units[@]} - 1 ))
	do
		bytes_value=$(echo "scale=3; $bytes_value / 1024" | bc -l)
		((i++))
	done

	value_align=$bytes_value
}

function show_available_memory() {
	local node=$1
	
	local allocatable=$(kubectl describe node $node | grep "Allocatable" -A6 | grep memory | awk '{print $2}')
	local allocated=$(kubectl describe node $node | grep "Allocated resources" -A8 | grep memory | awk '{print $2}')

	local unit_allocatable mem_bytes_allocatable
	local unit_allocated mem_bytes_allocated
	local mem_align_allocatable mem_align_allocated mem_align_available

	to_bytes $allocatable unit_allocatable mem_bytes_allocatable
	to_bytes $allocated unit_allocated mem_bytes_allocated
	local mem_bytes_available=$((mem_bytes_allocatable - mem_bytes_allocated))

	## Wants to align with higher unit, however, just use "GiB")
	align_unit $mem_bytes_allocatable "GiB" mem_align_allocatable
	align_unit $mem_bytes_allocated "GiB" mem_align_allocated
	align_unit $mem_bytes_available "GiB" mem_align_available

	printf "[Memory Status]\n"
	printf "<$node>\n"
	printf "%-11s : %8s %3s\n" "Allocatable" "$mem_align_allocatable" "GiB"
	printf "%-11s : %8s %3s\n" "Allocated" "$mem_align_allocated" "GiB"
	printf "%-11s : %8s %3s\n" "Available" "$mem_align_available" "GiB"
}

if [[ "$#" == "0" ]]; then
	usage
	exit 1
fi

function show_memory_all() {
	nodes=$(kubectl get node | awk '{print$1}' | sed '1d')
	arr=($nodes)
	printf "\n[All Worker Nodes in k8s Cluster]\n"
	for i in "${arr[@]}"; do
		if [[ $i == "Lab-K8sM" ]]; then
			continue
		fi
		printf "%s " "$i"
	done
	printf "\n"
	
	for i in "${arr[@]}"; do
		if [[ $i == "Lab-K8sM" ]]; then
			continue
		fi
		show_available_memory $i
	done
}

if [[ "$1" == "all" ]]; then
	show_memory_all
else
	nodes=$(kubectl get node | awk '{print$1}' | sed '1d')
	arr=($nodes)
	for i in "${arr[@]}"; do
		if [[ $i == "Lab-K8sM" ]]; then
			continue
		elif [[ $i == $1 ]]; then
			show_available_memory $1
			exit 0
		fi
	done

	printf "Could not find $1 from the k8s cluster.\n"
	printf "[All Worker Nodes in k8s Cluster]\n"
	for i in "${arr[@]}"; do
		if [[ $i == "Lab-K8sM" ]]; then
			continue
		fi
		printf "%s " "$i"
	done
	printf "\n"
	exit 1
fi
