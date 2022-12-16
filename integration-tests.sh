#!/bin/bash

# Note: this isnt a complete integration-test, this can be if adding jq to parse responses and comparing results
# however this is not a requirement.

#make start-kind
#helm plugin install https://github.com/jkroepke/helm-secrets
kubectl port-forward svc/portal-server 8443:8443 &

helm secrets upgrade portal-server helm-chart --install --create-namespace --namespace portal -f helm-charts/secrets.yaml

curl --request GET -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem https://localhos:8443/api/v1/namespaces/deployments
curl --request GET -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem https://localhost:8443/api/v1/namespaces/deployments
NAMESPACE=portal; curl --request GET -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem "https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments"
NAMESPACE=portal; NAME=portal-server; curl --request GET -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem "https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}"
NAMESPACE=portal; NAME=portal-server; REPLICAS=4; curl --request PUT -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem "https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}/replicas/${REPLICAS}"
NAMESPACE=portal; NAME=portal-server; REPLICAS=2; curl --request PUT -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem "https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}/replicas/${REPLICAS}/reconcile"
kubeclt scale deployment portal-server -n portal --replicas=4
kubectl scale deployment portal-server -n portal --replicas=4
NAMESPACE=portal; NAME=portal-server; curl --request GET -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem "https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}/diff"
kubectl scale deployment portal-server -n portal --replicas=4
NAMESPACE=portal; NAME=portal-server; curl --request GET -k --cert certs/client-1.crt --key certs/client-1.key --cacert certs/ca_cert.pem "https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}/diff"
