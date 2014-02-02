#!/bin/bash

set -e
set -x

ln -svf /vagrant/.vagrant-skel/bashrc ~/.bashrc
ln -svf /vagrant/.vagrant-skel/bash_profile ~/.bash_profile

source ~/.bashrc

go get -x github.com/kr/godep
go get -x code.google.com/p/go.tools/cmd/cover

curl -L https://get.rvm.io | bash -s stable --ruby=2.0.0
source ~/.bash_profile
gem install bunder foreman
