#! /bin/bash
set -e

CLIENT="client-1"
OUT_FOLDER="./certs"
DAYS=30
#KEY_CIPHER="aes256" this would force a passphrase and encrypt the key with aes256
CERT_CIPHER="sha256"
SUBJ="/C=US/ST=California/L=Mountain View/O=Your Organization/OU=Your Unit/CN=localhost"

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

print "Generating client key with"
openssl genrsa -out "${OUT_FOLDER}/${CLIENT}.key" 4096

print "Generating the client certificate signing request using ${CERT_CIPHER}"
openssl req -new -key "${OUT_FOLDER}/${CLIENT}.key" -days ${DAYS} -${CERT_CIPHER} -out "${OUT_FOLDER}/${CLIENT}.csr" -subj "${SUBJ}"

client_csr_signature=$(openssl req -text -noout -verify -in ${OUT_FOLDER}/${CLIENT}.csr | grep 'Signature Algorithm:' | head -1)
print "Client csr is using ${client_csr_signature}"

print "Signing the certificate with the Root CA"
openssl x509 -req -in "${OUT_FOLDER}/${CLIENT}.csr" \
    -extfile <(printf "subjectAltName=DNS:localhost") \
    -CA "${OUT_FOLDER}/ca_cert.pem" -CAkey "${OUT_FOLDER}/ca_key.pem"  \
    -days ${DAYS} -${CERT_CIPHER} -CAcreateserial \
    -out "${OUT_FOLDER}/${CLIENT}.crt"

client_certificate_signature=$(openssl x509 -in "${OUT_FOLDER}/${CLIENT}.crt" -text -noout | grep 'Signature Algorithm:' | head -1)
print "Client certificate is using ${client_certificate_signature}"

print_done "Done!"
