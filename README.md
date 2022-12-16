# portal
![Coverage](https://img.shields.io/badge/Coverage-61.1%25-yellow)
Kubernetes Replica API Server
[![Portal](./portal.jpeg)](https://goteleport.com)


# Design Doc
[RFD](./rfd/0000-kubernetes-replicas.md)

# Generate Certificates 

The following make targets will prompt you for a passphrase on the CA key.
use these target to generate client and server certificates

```bash
make certs-gen-server
make certs-gen-client
```

# How to run the server
Before you start, you will need to clone this repo and generate server and client certifictaes using make
```bash
make certs-gen-server
make certs-gen-client
```

Download the latest release from (releases)[https://github.com/innovia/portal/releases]

Run with the following command
```bash
./portal --tls_cert_file ./certs/server.crt --tls_private_key_file ./certs/server.key --ca_cert_file ./certs/ca_cert.pem --kubeconfig=/full/path/to/kubeconfig
```

## Graceful termination on OS signals
The server and reconcile loop would be able to handle a sig TERM or sig INT and gracefully shutdown

```text
I0830 17:32:34.038334   76627 main.go:79] Server started and listening on https://localhost:8443
I0830 17:32:38.112927   76627 reconcile.go:78] reconcile: nginx-deployment.default - drift detected => reconcile replicas 4 => 4
I0830 17:32:38.646391   76627 reconcile.go:25] Starting Replica Reconcile Loop...
^CI0830 17:32:41.040348   76627 reconcile.go:220] Stopping Reconcile Loop
I0830 17:32:41.040347   76627 main.go:87] Shutting down server
```

# Creating a release (automated)
before you run the following command, make sure git branch is clean.
make sure the version has the `v` perfix

```bash
VERSION=v1.0.0 make release/tag 
```

this will kick the GoReleaser and create the releases.

# Creating local binaries without docker push (not automated):
set the version to your requested version without the `v` prefix
```bash
VERSION=1.0.0 make release/tag 
make release
``` 

result binaries would be stored in the `dist` folder


# CI
github actions will run on every commit with:
* linting ([staticcheck](https://staticcheck.io/))
* unit-testing using goSumTests
* code-coverage - will create a badge
* build - code will be built

# Helm
helm secrets can be secured with gpg or kms, with encrypted secrets, this solution can be extended to github actions

this helm chart is production ready, with pod disruption budget and node anti-affinity, and 2 replicas to make sure the service is up.
it also has minimum required RBAC permissions for teh service account

configmap state will be auto created in the namespace of the install by the server on run.

health-checks for the pods are checking connectivity with kubernetes API server by making a raw request and receiving data back.
the health-checks are not mTLS since kubelet does not have the client certificates to communicate with the server, thus health-checks runs on another port (configurable)

[for more info read the helm-chart readme](helm-chart/README.md) 

# Client Curl Commands:
* [Curl Commands](./docs/client_curl_commands.md)

# Make Targets:
```console 
build              - build the server
release/tag        - use VERSION=v0.0.0 release/tag to create a new tag and push - this will trigger a GoReleaser build binaries and Docker push
release/dry-run    - run GoReleaser locally and skip pushing Docker images
release            - run GoReleaser locally and push to Docker images
test               - run unit tests
certs-cleanup      - delete certs folder
certs-gen-server   - generate server certificates and root CA
certs-gen-client   - generate client certificates
start-kind         - Run Kubernetes In Docker (Kind)
run-server-in-kind - run Portal server in Docker and network into kind
clean              - delete bin dist and certs folders
```
