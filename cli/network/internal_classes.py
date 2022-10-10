import subprocess
import json
import sys
import os

import invoke


class FileLocations:
    def __init__(self, network_config_location:str = "./configs/networks.json"):
        if os.path.exists(network_config_location):
            self.network_config_location = network_config_location
        else:
            print("Sorry, the network file config was not found here: " + network_config_location, file=sys.stderr)
            sys.exit(1)

        self.network_config_location = network_config_location
        
        with open(self.network_config_location, "r") as file:
            network_config_location_dict = file.read()
        
        self.network_config_location_dict = json.loads(network_config_location_dict)


class NetworkInit:
    def __init__(self):
        self.network_config_location_dict = FileLocations().network_config_location_dict
    
    def init(self):
        for network in self.network_config_location_dict["networks"]:
            
            network_name = network["bridge_name"]
            
            command = "ifconfig | grep -c vm-" + network_name + " || true"
            output = subprocess.check_output(command, shell=True)
            output = output.decode("utf-8").split()[0]
            
            if output != "1":
                command = "ifconfig bridge create name vm-" + network_name
                subprocess.run(command, shell=True, stdout=subprocess.DEVNULL)
                print(" 🔷 DEBUG: " + command)
                
                if network["bridge_interface"] and network["bridge_interface"] != "None":
                    command = "ifconfig vm-external addm " + network["bridge_interface"]
                    subprocess.run(command, shell=True, stdout=subprocess.DEVNULL)
                    print(" 🔷 DEBUG: " + command)

                if network["apply_bridge_address"] == True:
                    command = "ifconfig vm-" +  network_name + " inet " + network["bridge_address"] + "/" + str(network["bridge_subnet"])
                    subprocess.run(command, shell=True, stdout=subprocess.DEVNULL)
                    print(" 🔷 DEBUG: " + command)
            
            elif output == "1":
                print(" 🔷 DEBUG: Network " + network_name + " is already configured!")
            
            else:
                print(" 🚫 ERROR: Something unexpected happened!")
