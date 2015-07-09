#!/usr/bin/env bash

set -ex

cd `dirname $0`

if [ -f "$1" ]; then
  CONFIG_FILE=$1
else
  CONFIG_FILE=`mktemp -t watsXXXXX`
  trap "rm -f $CONFIG_FILE" EXIT

: ${API:?"Must set api url (e.g. api.10.244.0.34.xip.io)"}
: ${ADMIN_USER:?"Must set admin username (e.g. admin)"}
: ${ADMIN_PASSWORD:?"Must set admin password (e.g. admin)"}
: ${APPS_DOMAIN:?"Must set app domain url (e.g. 10.244.0.34.xip.io)"}
: ${SOCKET_ADDRESS_FOR_SECURITY_GROUP_TEST:?"Must set address [ip address of Diego ETCD cluster] (e.g. 10.244.16.2:4001)"}
: ${DOPPLER_URL:?"Must set doppler websocket url (e.g. wss://doppler.hello.cf-app.com:4443)"}

cat > $CONFIG_FILE <<HERE
{
  "api": "$API",
  "admin_user": "$ADMIN_USER",
  "admin_password": "$ADMIN_PASSWORD",
  "apps_domain": "$APPS_DOMAIN",
  "secure_address": "$SOCKET_ADDRESS_FOR_SECURITY_GROUP_TEST",
  "skip_ssl_validation": true
}
HERE
fi

if [[ "$(uname)" = "Darwin" ]]; then
  ln -sf cf-darwin ../bin/cf
else
  ln -sf cf-linux ../bin/cf
fi

export PATH=$PWD/../bin:$PATH

export GOPATH=$PWD/../../../../../
export GOBIN=$GOPATH/bin
export PATH=$GOBIN:$PATH

go install github.com/onsi/ginkgo/ginkgo
# The following line will fail with the || echo, since tests don't
# have a binary and go get will try to build one
go install -t ../tests/... 2>/dev/null || echo "Installed dependencies"

CONFIG=$CONFIG_FILE ginkgo -r -v -failFast -slowSpecThreshold=120 $@ ../wats
