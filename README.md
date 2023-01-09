## Migrations
Use [migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) to run migrations.

## Creating test certificates
```shell
openssl genrsa -out ca.key 4096
openssl req -new -sha512 -x509 -days 3650 -key ca.key -subj "/C=CN/ST=GD/L=SZ/O=Home Ltd./CN=Home Root CA" -out ca.crt
openssl req -newkey rsa:4096 -nodes -keyout test.key -subj "/C=CN/ST=GD/L=SZ/O=Home Ltd." -out server.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:*") -days 3650 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -sha512 -out test.crt
```

## Create sign key
```shell
openssl rand 64 > sign.key
```