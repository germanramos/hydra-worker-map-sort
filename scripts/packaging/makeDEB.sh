#!/bin/bash

### http://linuxconfig.org/easy-way-to-create-a-debian-package-and-local-package-repository

rm -rf ~/debbuild
mkdir -p ~/debbuild/DEBIAN
cp control ~/debbuild/DEBIAN

mkdir -p ~/debbuild/etc/hydra
cp ./fixtures/hydra-worker-map-sort.conf ~/debbuild/etc/hydra

mkdir -p ~/debbuild/etc/init.d
cp hydra-worker-map-sort-init.d.sh ~/debbuild/etc/init.d/hydra-worker-map-sort

mkdir -p ~/debbuild/usr/local/hydra
cp ../../bin/hydra-worker-map-sort  ~/debbuild/usr/local/hydra

chmod -R 644 ~/debbuild/usr/local/hydra/*
chmod 755 ~/debbuild/etc/init.d/hydra-worker-map-sort
chmod 755 ~/debbuild/usr/local/hydra/hydra-worker-map-sort

sudo chown -R root:root ~/debbuild/*

pushd ~
sudo dpkg-deb --build debbuild

popd
sudo mv ~/debbuild.deb hydra-worker-map-sort-1-1.x86_64.deb
