#!/usr/bin/env sh

set -e

echo "Staring the update process..."
echo ""

echo "Pulling updates from Git..."
git pull
echo ""

echo "Building the hoster module..."
go build
echo "Done"
echo ""

echo "Building the vm_supervisor module..."
cd vm_supervisor/
go build
mv vm_supervisor ../vm_supervisor_service
echo "Done"
echo ""

echo "Done building!"
