
NAMESPACE=ddis-asteroid--usernaut-${ENVIRONMENT}

kubectl create cm usernaut-config --from-file=default.yaml=./appconfig/default.yaml \
--from-file=${ENVIRONMENT}.yaml=./appconfig/${ENVIRONMENT}.yaml -n ${NAMESPACE}

