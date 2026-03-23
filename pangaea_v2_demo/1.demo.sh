#!/bin/bash
# more pods on CXL Host than Normal Host

readonly SCRIPT_PATH=$(dirname $(readlink -f "${BASH_SOURCE[0]}"))
readonly APPLY_SCRIPT="$SCRIPT_PATH/apply_pod.sh"
readonly DELETE_SCRIPT="$SCRIPT_PATH/delete_pod.sh"
readonly lspci_cmd="lspci -tv -s 20:01.1;echo"
readonly lsmem_cmd="lsmem --output-all;lsmem --output-all;echo"
readonly WORKER_NODE_IP="192.168.0.11"

function pause(){
	read -n 1 -s -r
}

function apply_pod(){
	$APPLY_SCRIPT $1 $2 silent
}

function delete_pod(){
	$DELETE_SCRIPT $1 $2 silent force
}

function show_lspci(){
	IFS=$'\n' RESULT=(`ssh -t root@${WORKER_NODE_IP} $lspci_cmd 2>/dev/null`)
	echo \< CXL Device PCI Tree \>; 
	for VALUE in "${RESULT[@]}"
	do
		echo "$VALUE"
		if [[ "$VALUE" == *Montage* ]]; then
			echo "                                                                Samsung Electronics Co Ltd Device 0112"
		fi
	done
}

NODE=gigabyte-turin
PODNUM=5

clear

printf "[CXL VCS Bind Status - initial]\n"
show_lspci

pause
clear

printf "[Apply $PODNUM pods - each pod requests 60 GiB memory]\n"
for ((i=1;i<=$PODNUM;i++))
do
	pause
	printf "# kubectl apply %s\n\n" "python-alloc-60gb-$i"
	apply_pod $NODE $i
	if [ $((i %2)) -eq 1 ]; then
		printf "Bind additional CXL Memory Device...\n"
		pause
		printf "[CXL VCS Bind Status - after bind]\n"
		show_lspci
		continue
	fi
done

pause
clear

printf "[Delete deployed pods]\n"
for ((i=$PODNUM;i>=1;i--))
do
	pause
	printf "# kubectl delete %s\n\n" "python-alloc-60gb-$i"
	delete_pod $NODE $i
	if [ $((i %2)) -eq 1 ]; then
		pause
		printf "Unbind unused CXL Memory Device...\n"
		pause
		printf "[CXL VCS Bind Status - after unbind]\n"
		show_lspci
		continue
	fi
done
