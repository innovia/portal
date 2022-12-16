# portal

This chart will install a portal replicas server with reconcile 

Note: you may choose not to encrypt the certificates and server key, in this case simply fill in the values for the `mtls.server_key` `mtls.server_cert` and `ca_cert` then run the following command (replace the values file path with the path to your new values file):

```bash
helm upgrade portal-server --install --create-namespace --namespace portal -f  <path/to/your/values/file>  helm-chart
```

The instructions below document a secure approach to storing secrets in git and encrypting/decrypting for deployment

## Generate a PGP key and encrypt secrets.yaml file with the content of, or use this as the values file with the content of the certs from the certs folder

```yaml
mtls:
  server_key: |
    <content of server.key here>
  server_cert: |
    <content of server.crt here>
  ca_cert: |
    <content of ca_cert.pem here>
 ```

## Before installing the Chart (only use if you are using gpg)
You will need to install:
* [gpg](https://formulae.brew.sh/formula/gnupg)
* [sops](https://github.com/mozilla/sops/tree/0494bc41911bc6e050ddd8a5da2bbb071a79a5b7#up-and-running-in-60-seconds)
* [helm-secrets plugin](https://github.com/jkroepke/helm-secrets) 

### Generating PGP Key

```bash
 gpg --full-generate-key
```

From the list of GPG keys, copy the long form of the GPG key ID:

```bash
gpg --list-secret-keys --keyid-format=long
```

substitute in the GPG key ID you'd like to use.

```bash
gpg --export-secret-keys --armor {{fingerprint}} > private.rsa
# Prints the GPG key ID, in ASCII armor format
```

add to gpg keychain
```bash
gpg --import private.rsa
```

update the `.sops.yaml` file with your key fingerprint

### Encrypt secrets.yaml

```bash
export SOPS_PGP_FP='{{full_fingerprint}}' 
sops -e -i secrets.yaml
```


## Installing the Chart

 
```bash
$ helm secrets --install  --install --create-namespace portal  portal .
```


## Configuration

The following tables lists configurable parameters of the portal chart and their default values:

| Parameter                        | Description                                                                  | Default                              |
|----------------------------------|------------------------------------------------------------------------------|--------------------------------------|
| affinity                         | affinities to use                                                            | `{}`                                 |
| image.pullPolicy                 | image pull policy                                                            | `IfNotPresent`                       |
| image.repository                 | image repo that contains the admission server                                | `innovia/portal`                     |
| image.tag                        | image tag                                                                    | `1.0.0`                              |
| image.imagePullSecrets           | image pull secrets for private repositories                                  | `[]`                                 |
| namespaceSelector                | namespace selector to use, will limit webhook scope                          | `{}`                                 |
| nodeSelector                     | node selector to use                                                         | `{}`                                 |
| podAnnotations                   | extra annotations to add to pod metadata                                     | `{}`                                 |
| replicaCount                     | number of replicas                                                           | `2`                                  |
| resources                        | resources to request                                                         | `{}`                                 |
| service.externalPort             | webhook service external port                                                | `443`                                |
| service.name                     | webhook service name                                                         | `portal`                             |
| service.type                     | webhook service type                                                         | `ClusterIP`                          |
| tolerations                      | tolerations to add                                                           | `[]`                                 |
| rbac.enabled                     | use rbac                                                                     | `true`                               |
| volumes                          | extra volume definitions                                                     | `[]`                                 |
| volumeMounts                     | extra volume mounts                                                          | `[]`                                 |
| podDisruptionBudget.enabled      | enable PodDisruptionBudget                                                   | `false`                              |
| podDisruptionBudget.minAvailable | represents the number of Pods that must be available (integer or percentage) | `1`                                  |





#
