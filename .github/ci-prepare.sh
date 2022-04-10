#!/bin/bash
sudo apt-get install -y software-properties-common
sudo apt-key adv --fetch-keys 'https://mariadb.org/mariadb_release_signing_key.asc'
sudo add-apt-repository 'deb [arch=amd64,arm64,ppc64el] https://mirrors.bkns.vn/mariadb/repo/10.6/ubuntu focal main'
sudo apt-get update
sudo apt-get install -f mariadb-client
