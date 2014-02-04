#!/bin/bash

export DEBIAN_FRONTEND=noninteractive

umask 022

set -e
set -x

apt-get update -yq
apt-get install -yq \
  build-essential \
  byobu \
  bzr \
  curl \
  git \
  make \
  mercurial \
  ruby1.9.3 \
  screen

if ! go env ; then
  curl -s -L https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz | \
    tar xzf - -C /usr/local
  ln -svf /usr/local/go/bin/* /usr/local/bin/
fi

if ! docker version ; then
  curl -s https://get.docker.io | sh
fi

mkdir -p /gopath
GOPATH=/gopath go get -x code.google.com/p/go.tools/cmd/cover
chown -R vagrant:vagrant /gopath

su - vagrant -c /vagrant/.vagrant-provision-as-vagrant.sh
