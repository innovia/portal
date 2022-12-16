
# Client Curl Commands:

* [Get deployments all namespaces](#get-deployments)
* [Get deployments from a namespace](#get-deployments-from-a-namespace)
* [Get a deployment from a namespace](#get-deployment-from-namespace)
* [Set replicas for a deployment](#set-replicas-for-a-deployment)
* [Set replicas and reconcile for a deployment](#set-replicas-and-reconcile-for-a-deployment)
* [Show replicas diff for a deployment](#show-replicas-diff-for-a-deployment)
* [Kubernetes API health check](#kubernetes-api-health-check)

<hr/>

### Get deployments

```bash
curl --request GET -k \
--cert certs/client-1.crt \
--key certs/client-1.key \
--cacert certs/ca_cert.pem \
https://localhost:8443/api/v1/namespaces/deployments

```
#### expected response
```text
HTTP/2 200 
content-type: application/json
content-length: 1563
date: Tue, 30 Aug 2022 04:23:22 GMT

{
    "count": 20,
    "deployments": [
        {
            "name": "<name>",
            "namespace": "<namepsace>",
            "replicas": 1
        }...
```

<hr/>

### Get deployments from a namespace

```bash
NAMESPACE=<namespace>

curl --request GET -k \
--cert certs/client-1.crt \
--key certs/client-1.key \
--cacert certs/ca_cert.pem \
"https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments"
```

#### expected response
```text
HTTP/2 200 
content-type: application/json
content-length: 1563
date: Tue, 30 Aug 2022 04:23:22 GMT

{
    "count": 2,
    "deployments": [
        {
            "name": "deployment-1",
            "namespace": "<namespace>",
            "replicas": 1
        },
        {
            "name": "deployment-2",
            "namespace": "<namespace>",
            "replicas": 1
        }
    ]
}
```

<hr/>

### Get deployment from namespace

```bash
NAMESPACE=<namespace>
NAME=<name>

curl --request GET -k \
--cert certs/client-1.crt \
--key certs/client-1.key \
--cacert certs/ca_cert.pem \
"https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}"
```

#### expected response
```text
HTTP/2 200 
content-type: application/json
content-length: 1563
date: Tue, 30 Aug 2022 04:23:22 GMT

{
    "name": "<name>",
    "namespace": "<namespace>",
    "replicas": 1
}
```

<hr/>

### Set replicas for a deployment

```bash
NAMESPACE=<namespace>
NAME=<name>
REPLICAS=<replicas>

curl --request PUT -k \
--cert certs/client-1.crt \
--key certs/client-1.key \
--cacert certs/ca_cert.pem \
"https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}/replicas/${REPLICAS}"
```

#### expected response
```text
HTTP/2 200 
content-type: application/json
content-length: 1563
date: Tue, 30 Aug 2022 04:23:22 GMT

{
    "name": "<name>",
    "namespace": "<namespace>",
    "replicas": <replicas>,
    "Reconcile": false,
    "Time": "2022-08-30T00:33:36.280228-04:00"
}
```

<hr/>

### Set replicas and reconcile for a deployment

```bash
NAMESPACE=<namespace>
NAME=<name>
REPLICAS=<replicas>

curl --request PUT -k \
--cert certs/client-1.crt \
--key certs/client-1.key \
--cacert certs/ca_cert.pem \
"https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}/replicas/${REPLICAS}/reconcile"
```

#### expected response
```text
HTTP/2 200 
content-type: application/json
content-length: 1563
date: Tue, 30 Aug 2022 04:23:22 GMT

{
    "name": "<name>",
    "namespace": "<namespace>",
    "replicas": <replicas>,
    "Reconcile": true,
    "Time": "2022-08-30T00:37:52.477146-04:00"
}
```

<hr/>

### Show replicas diff for a deployment

```bash
NAMESPACE=<namespace>
NAME=<name>

curl --request GET -k \
--cert certs/client-1.crt \
--key certs/client-1.key \
--cacert certs/ca_cert.pem \
"https://localhost:8443/api/v1/namespaces/${NAMESPACE}/deployments/${NAME}/diff"
```

#### expected response
```text
HTTP/2 200 
content-type: application/json
content-length: 1563
date: Tue, 30 Aug 2022 04:23:22 GMT


{
    "name": "<name>",
    "namespace": "<namespace>",
    "diff": "replicas: 2 => 1"
}
```

<hr/>

### Kubernetes API Health Check

```bash
curl --request GET -k \
--cert certs/client-1.crt \
--key certs/client-1.key \
--cacert certs/ca_cert.pem \
"https://localhost:8443/livez"
```

#### expected response
```text
HTTP/2 200 
content-type: application/json
content-length: 1563
date: Tue, 30 Aug 2022 04:23:22 GMT

ok
```
