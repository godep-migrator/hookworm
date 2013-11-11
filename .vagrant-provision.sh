#!/bin/bash

export DEBIAN_FRONTEND=noninteractive

umask 022

set -e
set -x

apt-get update -yq
apt-get install -yq curl make git-core

if [ ! -f /var/tmp/godeb ] ; then
  cd /var/tmp
  curl -s https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz | tar xzf -
  ./godeb install 1.1.2
fi

if ! which docker ; then
  curl -s https://get.docker.io | sh
fi
