#!/bin/bash

### http://tecadmin.net/create-rpm-of-your-own-script-in-centosredhat/#

sudo yum install rpm-build rpmdevtools
rm -rf ~/rpmbuild
rpmdev-setuptree

mkdir ~/rpmbuild/SOURCES/hydra-worker-map-sort-1
cp ./fixtures/hydra-worker-map-sort.conf  ~/rpmbuild/SOURCES/hydra-worker-map-sort-1
cp hydra-worker-map-sort-init.d.sh ~/rpmbuild/SOURCES/hydra-worker-map-sort-1
cp ../../bin/hydra-worker-map-sort ~/rpmbuild/SOURCES/hydra-worker-map-sort-1

cp hydra-worker-map-sort.spec ~/rpmbuild/SPECS

pushd ~/rpmbuild/SOURCES/
tar czf hydra-worker-map-sort-1.1.tar.gz hydra-worker-map-sort-1/
cd ~/rpmbuild 
rpmbuild -ba SPECS/hydra-worker-map-sort.spec

popd
cp ~/rpmbuild/RPMS/x86_64/hydra-worker-map-sort-1-1.x86_64.rpm .
