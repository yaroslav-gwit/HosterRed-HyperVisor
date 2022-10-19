import subprocess
import json
import sys
import os

from rich.console import Console
from rich.table import Table
from rich import box
import invoke
import typer


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



""" Section below is responsible for the CLI input/output """
app = typer.Typer(context_settings=dict(max_content_width=800))


@app.command()
def init():
    """ Initialize Hoster networks """

    NetworkInit().init()


@app.command()
def info(json_pretty:bool = typer.Option(False, help="Pretty JSON output"),
        json_plain:bool = typer.Option(False, help="Plain and compliant JSON output"),
        table_title:bool = typer.Option(False, help="Show table title"),
        table:bool = typer.Option(True, help="Table output"),
    ):

    """ Show host network infomation """

    network_config_dict = FileLocations().network_config_location_dict
    network_config_json = json.dumps(network_config_dict)

    if table:
        if not table_title:
            table = Table(box=box.ROUNDED, show_lines = True)
        else:
            table = Table(title = " Network Interfaces", box=box.ROUNDED, title_justify = "left", show_lines = True)

        table.add_column("#", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Name", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Address", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Subnet", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("IP Range", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Interface", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Address apply", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Comment", justify="center", style="bright_cyan", no_wrap=True)

        iteration = 0
        for item in network_config_dict["networks"]:
            iteration = iteration + 1
            table.add_row(
                str(iteration),
                item["bridge_name"],
                item["bridge_address"],
                str(item["bridge_subnet"]),
                str(item["range_start"]) + "-" + str(item["range_end"]),
                item["bridge_interface"],
                str(item["apply_bridge_address"]),
                item["comment"],
            )

        Console().print(table)

    elif json_plain:
        print(network_config_json)

    elif json_pretty:
        Console().print_json(network_config_json, indent = 3, sort_keys=False)
