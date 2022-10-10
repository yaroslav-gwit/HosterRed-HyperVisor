#!/usr/bin/env python3
import json
import os
import subprocess
import sys

class LoadKernelModules:
    def __init__(self):
        """
        kldstat -m $MODULE
        kldstat -mq $MODULE

        kldload vmm
        kldload nmdm
        kldload if_bridge
        kldload if_tuntap
        kldload if_tap

        sysctl net.link.tap.up_on_open=1

        13.0-RELEASE-p11
        """
    
    def init(self):
        command = "kldload vmm"
        print(" 🔷 DEBUG: " + command)
        subprocess.run(command, shell=True)

        command = "kldload nmdm"
        print(" 🔷 DEBUG: " + command)
        subprocess.run(command, shell=True)

        command = "kldload if_bridge"
        print(" 🔷 DEBUG: " + command)
        subprocess.run(command, shell=True)

        command = "kldload if_tuntap"
        print(" 🔷 DEBUG: " + command)
        subprocess.run(command, shell=True)

        command = "sysctl net.link.tap.up_on_open=1"
        print(" 🔷 DEBUG: " + command)
        subprocess.run(command, shell=True)
