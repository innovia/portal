#! /bin/bash
set -e

OUT_FOLDER="./certs"
KEY_CIPHER="aes256"
CERT_CIPHER="sha256"
DAYS=30
SUBJ="/C=US/ST=California/L=San Francisco/O=Teleport/OU=Cloud Tooling/CN=localhost"

RESET=$(tput sgr0)
COLOR_WHITE=$(tput bold setaf 7)
COLOR_BLUE=$(tput setaf 4)
COLOR_GREEN=$(tput setaf 2)

print(){
  printf "\n${COLOR_BLUE}==> ${RESET} ${COLOR_WHITE} %-20s ${RESET}\n" "${1}"
}

print_done(){
  printf "\n${COLOR_GREEN}==> ${RESET} ${COLOR_WHITE} %s${RESET}${COLOR_GREEN} ✔ ${RESET}\n" "${1}"
}

mkdir -p "${OUT_FOLDER}"

print "Generating Root CA Private Key encrypted with ${KEY_CIPHER} - valid for ${DAYS} days"
openssl genrsa -${KEY_CIPHER} -out "${OUT_FOLDER}/ca_key.pem" 4096

print "Generating Root CA Public Certificate using sha256 - valid for 5 year"
openssl req -key "${OUT_FOLDER}/ca_key.pem" -new -x509 -days 1825 -sha256  -subj "$SUBJ" -out "${OUT_FOLDER}/ca_cert.pem"

# check out public cert - The Signature Algorithm should be using SHA-256.
signature=$(openssl x509 -in ${OUT_FOLDER}/ca_cert.pem -text -noout | grep 'Signature Algorithm:' | head -1)
print "Root CA Certificate ${signature}"

print "Generating the server key"
openssl genrsa -out "${OUT_FOLDER}/server.key" 4096

print "Generating the server certificate signing request using ${CERT_CIPHER}"
openssl req -key "${OUT_FOLDER}/server.key" -new -days ${DAYS} -${CERT_CIPHER} -out "${OUT_FOLDER}/server.csr" -subj "${SUBJ}"

print "Signing the certificate with the Root CA"
server_csr_signature=$(openssl req -text -noout -verify -in ${OUT_FOLDER}/server.csr | grep 'Signature Algorithm:' | head -1)
print "Server csr is using ${server_csr_signature}"

openssl x509 -req -in "${OUT_FOLDER}/server.csr" \
    -extfile <(printf "subjectAltName=DNS:localhost") \
    -CA "${OUT_FOLDER}/ca_cert.pem" -CAkey "${OUT_FOLDER}/ca_key.pem"  \
    -days ${DAYS} -${CERT_CIPHER} -CAcreateserial \
    -out "${OUT_FOLDER}/server.crt"

server_certificate_signature=$(openssl x509 -in ${OUT_FOLDER}/server.crt -text -noout | grep 'Signature Algorithm:' | head -1)
print "Server certificate is using ${server_certificate_signature}"

print_done "Done!"
