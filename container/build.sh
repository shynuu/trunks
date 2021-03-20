#!/bin/bash

VERSION=$(curl -s https://raw.githubusercontent.com/shynuu/trunks/master/version.txt)
curl -o trunks -LO https://github.com/shynuu/trunks/releases/download/$VERSION/trunks_amd64
docker build -t trunks .