#!/bin/bash
set -e

cd $(dirname $0)/../generator

ID=$(curl -s 'http://localhost:8080/v1/accounts?uuid=service' | jq -r .data[0].id)
if [ "$ID" = "null" ]; then
    ID=$(curl -s -F uuid=service -F name=service -F kind=service http://localhost:8080/v1/accounts | jq .id)
    echo Created account $ID
fi

KEYS=$(curl -s 'http://localhost:8080/v1/apikeys?publicValue=service&_role=superadmin' | jq -r '.data | length')

if [ "$KEYS" = "0" ]; then
    curl -s -F publicValue=service -F secretValue=servicepass -F accountId=$ID 'http://localhost:8080/v1/apikeys' | jq .
    echo Created key service:servicepass
    sleep 2 # Wait for the key to be active
fi

echo Using account $ID

curl -s -u service:servicepass http://localhost:8080/v1/schemas | jq . > schemas.json
echo Saved schemas.json

echo -n Generating go code...
godep go run generator.go
echo " Done"

gofmt -w ../client/generated_*
echo Formatted code

echo Success
