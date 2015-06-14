#!/bin/bash

echo export RANCHER_URL=http://localhost:8080
echo export RANCHER_ACCESS_KEY=$(grep 'access-key=' tox.ini | cut -f2 -d'=')
echo export RANCHER_SECRET_KEY=$(grep 'secret-key=' tox.ini | cut -f2 -d'=')
