#!/bin/bash

echo -n "Certificates generated successfully"
echo " "
openssl genrsa -out ca.key 4096
openssl req -new -x509 -key ca.key -sha256 -subj "/C=IN/ST=TN/O=KubeArmor" -days 365 -out ca.cert
openssl genrsa -out service.key 4096
openssl req -new -key service.key -out service.csr -config certificate.conf
openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial -out service.pem -days 365 -sha256 -extfile certificate.conf -extensions req_ext

echo -n "Attaching certificates to k8s secrets"
echo " "
kubectl create secret tls server-certs --cert=service.pem  --key=service.key  -n kube-system
kubectl create secret generic ca-cert --from-file=ca.cert -n kube-system

echo -n "Removing local copy of certs"
echo " "
rm -rf ca.* service.*
