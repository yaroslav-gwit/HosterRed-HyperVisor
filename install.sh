#!/usr/bin/env sh
set -e

bash build.sh

mkdir /opt/hoster-core/
cp hoster /opt/hoster-core/
cp vm_supervisor_service /opt/hoster-core/
cp -r config_files /opt/hoster-core/
ln -s /opt/hoster-core/hoster /bin/hoster

echo "Done installing"
