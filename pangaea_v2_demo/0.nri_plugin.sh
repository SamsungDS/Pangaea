#!/bin/bash
# install / uninstall CXL NRI Plugin

readonly SCRIPT_PATH=$(dirname $(readlink -f "${BASH_SOURCE[0]}"))
readonly PANGAEA_PATH="${SCRIPT_PATH}/../"
readonly PLUGIN_PATH="${PANGAEA_PATH}/CXL_NRI_Plugin"

function usage() {
    echo "<Usage>
        $0 <command>

        <command>                              <description>

        --install(-i)                          install nri plugin.
        --uninstall(-u)                        uninstall nri plugin.
    "
    exit 1
}

function install_plugin(){
        echo "install nri plugin."

	helm install cxl $PLUGIN_PATH/deployment/helm/cxl -n kube-system -f $PLUGIN_PATH/deployment/helm/cxl/values.yaml
}

function uninstall_plugin(){
        echo "uninstall nri plugin."

	helm uninstall cxl -n kube-system
}


if [ "$#" -ne 0 ]; then
    case "$1" in
        "--install"|"-i")
            install_plugin
            ;;
        "--uninstall"|"-u")
            uninstall_plugin
            ;;
        *)
	    usage
            ;;
    esac
else
    usage
fi

