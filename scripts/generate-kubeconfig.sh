#!/bin/bash

K8S_DIR=""
function get_path {
    DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    K8S_DIR=$(dirname $DIR)/artifacts/kubernetes
}
function print_error {
  echo && echo 'An error occurred while deploying'
  exit 1
}
function print_success {
  echo && echo 'CN-WAN Operator deployed successfully'
  exit 0
}
trap print_error ERR

function print_help {
    echo "Usage:"
    echo "generate-kubeconfig.sh [options]"
    echo
    echo "Generates a kubeconfig file to use for the cnwan-reader, based on the Kubernetes service account"
    echo "and cluster role as included in this repository."
    echo 
    echo "Options:"
    echo "--kubeconfig  path to the kubeconfig file to use. If empty, the default one will be used."
    echo "--namespace   the namespace where the Kubernetes service account file has been deployed to (default: default)."
    echo "--context     the context to use. If empty, the one being used with kubectl will be used."
    echo "--output      path to the file to create. If empty, a config file in the current directory will be created."
    echo "--help        show this help."
    echo
    echo "Examples:"
    echo "generate-kubeconfig.sh --context gke --output /path/to/config"
    echo "generate-kubeconfig.sh --kubeconfig /another/.kube/config"
    echo
}

SERVICE_ACCOUNT_NAME=cnwan-reader-service-account
CONTEXT=$(kubectl config current-context)
NAMESPACE=default
NEW_CONTEXT=cnwan-reader
BASE_KUBECONFIG=""
NEW_KUBECONFIG=$(pwd)/config

get_path

while test $# -gt 0; do
    case "$1" in

        --help)
            print_help
            exit 0
        ;;

        --kubeconfig)
            shift
            if test $# -gt 0; then
                BASE_KUBECONFIG="$(echo --kubeconfig) $1"
            fi
            shift
        ;;

        --namespace)
            shift
            if test $# -gt 0; then
                NAMESPACE=$1
            fi
            shift
        ;;

        --context)
            shift
            if test $# -gt 0; then
                CONTEXT=$1
            fi
            shift
        ;;

        --output)
            shift
            if test $# -gt 0; then
                NEW_KUBECONFIG=$1
            fi
            shift
        ;;

        *)
            break
        ;;
    esac
done

touch $NEW_KUBECONFIG

SECRET_NAME=$(kubectl ${BASE_KUBECONFIG} get serviceaccount ${SERVICE_ACCOUNT_NAME} \
  --context ${CONTEXT} \
  --namespace ${NAMESPACE} \
  -o jsonpath='{.secrets[0].name}')
TOKEN_DATA=$(kubectl ${BASE_KUBECONFIG} get secret ${SECRET_NAME} \
  --context ${CONTEXT} \
  --namespace ${NAMESPACE} \
  -o jsonpath='{.data.token}')

TOKEN=$(echo ${TOKEN_DATA} | base64 -d)

kubectl ${BASE_KUBECONFIG} config view --raw > ${NEW_KUBECONFIG}
kubectl --kubeconfig ${NEW_KUBECONFIG} config use-context ${CONTEXT}
kubectl --kubeconfig ${NEW_KUBECONFIG} config view --flatten --minify > ${NEW_KUBECONFIG}.tmp
kubectl config --kubeconfig ${NEW_KUBECONFIG}.tmp rename-context ${CONTEXT} ${NEW_CONTEXT}
kubectl config --kubeconfig ${NEW_KUBECONFIG}.tmp set-credentials ${CONTEXT}-${NAMESPACE}-token-user --token ${TOKEN}
kubectl config --kubeconfig ${NEW_KUBECONFIG}.tmp set-context ${NEW_CONTEXT} --user ${CONTEXT}-${NAMESPACE}-token-user
kubectl config --kubeconfig ${NEW_KUBECONFIG}.tmp set-context ${NEW_CONTEXT} --namespace ${NAMESPACE}
kubectl config --kubeconfig ${NEW_KUBECONFIG}.tmp view --flatten --minify > ${NEW_KUBECONFIG}
rm ${NEW_KUBECONFIG}.tmp