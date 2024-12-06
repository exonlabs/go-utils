#!/bin/bash
cd $(dirname $(readlink -f $0))

# clean all
rm -f *.pem

# RootCA
echo -e "\nROOT CA:\n"
openssl ecparam -name prime256v1 -genkey -out ca-key.pem
openssl req -new -x509 -days 3650 \
   -key ca-key.pem -out ca-cert.pem \
   -subj "/C=EG/O=ExonLabs/CN=RootCA" \
   -addext "keyUsage=critical,keyCertSign,cRLSign"
cat ca-key.pem ca-cert.pem


# Server
echo -e "\nSERVER:\n"
openssl ecparam -name prime256v1 -genkey -out server-key.pem
openssl req -new \
   -key server-key.pem -out server-csr.pem \
   -subj "/C=EG/O=ExonLabs/CN=server1"
openssl x509 -req -days 3650 \
   -in server-csr.pem -out server-cert.pem \
   -CA ca-cert.pem -CAkey ca-key.pem \
   -extfile <(printf " \
      \nkeyUsage=critical,digitalSignature,keyAgreement,keyEncipherment \
      \nextendedKeyUsage=serverAuth \
      \nsubjectAltName=DNS.1:server1.local,DNS.2:server1.lan,IP:127.0.0.1")
cat server-key.pem server-cert.pem

# Client
echo -e "\nCLIENT:\n"
openssl ecparam -name prime256v1 -genkey -out client-key.pem
openssl req -new \
   -key client-key.pem -out client-csr.pem \
   -subj "/C=EG/O=ExonLabs/CN=client1"
openssl x509 -req -days 3650 \
   -in client-csr.pem -out client-cert.pem \
   -CA ca-cert.pem -CAkey ca-key.pem \
   -extfile <(printf " \
      \nkeyUsage=critical,digitalSignature,keyAgreement,keyEncipherment \
      \nextendedKeyUsage=clientAuth \
      \nsubjectAltName=DNS.1:client1.local,DNS.2:client1.lan,IP.1:127.0.0.1")
cat client-key.pem client-cert.pem

echo -e "\n"
