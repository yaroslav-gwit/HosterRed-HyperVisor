#!/usr/bin/env sh
set -e

bash pull_changes.sh
echo ""

bash build.sh
echo ""
echo "=== Starting the installation process ==="

mkdir -p /opt/hoster-core/

cp hoster /opt/hoster-core/
cp vm_supervisor_service /opt/hoster-core/
cp -r config_files /opt/hoster-core/

rm -f /bin/hoster
ln -s /opt/hoster-core/hoster /bin/hoster

echo "=== Installation process done ==="
