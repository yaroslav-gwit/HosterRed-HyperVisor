#!/usr/bin/env sh
go build
cd vm_supervisor/
go build
mv vm_supervisor ../vm_supervisor_service
echo "Done building!"
