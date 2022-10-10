# system imports
from os.path import exists
import sys
import json
import re
import subprocess

# 3rd party imports
from generate_mac import generate_mac
import typer
from natsort import natsorted



# Random generators section -> Passwords, IPs and MACs
def random_pw_gn(capitals:bool = False, numbers:bool = False, lenght:int = 8, specials:bool = False):
    letters_var = "asdfghjklqwertyuiopzxcvbnm"
    capitals_var = "ASDFGHJKLZXCVBNMQWERTYUIOP"
    numbers_var = "0987654321"
    specials_var = "-_!?><)([]@"

    valid_chars_list = []
    for item in letters_var:
        valid_chars_list.append(item)
    if capitals:
        for c_item in capitals_var:
            valid_chars_list.append(c_item)
    if numbers:
        for n_item in numbers_var:
            valid_chars_list.append(n_item)
    if specials:
        for s_item in specials_var:
            valid_chars_list.append(s_item)

    password = ""
    for _i in range(0, lenght):
        password = password + random.choice(valid_chars_list)

    return password


def mac_address_generator(prefix:str = "58:9C:FC"):
    mac_addess = generate_mac.vid_provided(prefix)
    mac_addess = mac_addess.lower()

    return mac_addess


def ip_address_generator(ip_address:str = "random"):
    existing_ip_addresses = []
    for _vm in VmList().plainList:
        ip_address = CoreChecks(vm_name=_vm).vm_ip_address()
        existing_ip_addresses.append(ip_address)
    return existing_ip_addresses

    with open("./configs/networks.json", "r") as file:
        networks_file = file.read()
    networks_dict = json.loads(networks_file)
    networks = networks_dict["networks"][0]

    if ip_address in existing_ip_addresses:
        print("VM with such IP exists: " + ip_address)

    elif ip_address == "random":
        bridge_address = networks["bridge_address"]
        range_start = networks["range_start"]
        range_end = networks["range_end"]

        # Generate full list of IPs for the specified range
        bridge_split = bridge_address.split(".")
        del bridge_split[-1]
        bridge_join = ".".join(bridge_split) + "."

        ip_address_list = []
        for number in range(range_start, range_end+1):
            _ip_address = bridge_join + str(number)
            ip_address_list.append(_ip_address)

        ip_address = ip_address_list[0]
        number = range_start
        while ip_address in existing_ip_addresses:
            number = number + 1
            if number > range_end:
                sys.exit("There are no free IPs left!")
            else:
                ip_address = bridge_join + str(number)

    return ip_address


# VM CORE CHECKS
def vm_live(vm_name:str) -> bool:
    if exists("/dev/vmm/" + vm_name):
        return True
    else:
        return False


glob_zfs_datasets = {}
def dataset_list(ds_db_file:str = "./configs/datasets.json") -> dict:
    if not glob_zfs_datasets:
        with open(ds_db_file, "r") as file:
            datasets_file = file.read()
        glob_zfs_datasets = json.loads(datasets_file)

    return glob_zfs_datasets


glob_vm_info_dict = {}
glob_conf_location = "/vm_config.json"
def vm_config_read(vm_name:str, conf_location:str = "/vm_config.json") -> dict:
    if glob_vm_info_dict and glob_vm_info_dict.get("vm_name", "") == vm_name:
        return glob_vm_info_dict
    else:
        zfs_datasets = dataset_list()

        for ds in zfs_datasets["datasets"]:
            vm_config = ds["mount_path"] + vm_name + conf_location
            if exists(vm_config):
                with open(vm_config, 'r') as file:
                    vm_info_raw = file.read()
                glob_vm_info_dict = json.loads(vm_info_raw)
                glob_vm_info_dict["vm_name"] = vm_name
                glob_conf_location = conf_location
                
                return glob_vm_info_dict

            elif ds == zfs_datasets["datasets"][-1] and not exists(vm_config):
                print("Sorry, config file was not found for " + vm_name + " path: " + vm_config)
                sys.exit(1)


# VM STORAGE CHECKS
def vm_encrypted(vm_name:str) -> bool:
    zfs_datasets = dataset_list()

    result = False
    for ds in zfs_datasets["datasets"]:
        if exists(ds["mount_path"] + vm_name):
            if ds["encrypted"]:
                result = True

    return result


def vm_disk_exists(vm_name:str, disk_image_name:str) -> bool:
    zfs_datasets = dataset_list()

    result = False
    for ds in zfs_datasets["datasets"]:
        if exists(ds["mount_path"] + vm_name + "/" + disk_image_name):
            result = True

    return result


def disk_location(vm_name:str, disk_image_name:str) -> str:
    zfs_datasets = dataset_list()
    
    for ds in zfs_datasets["datasets"]:
        image_path = ds["mount_path"] + vm_name + "/" + disk_image_name
        if exists(image_path):
            return image_path


def vm_location(vm_name:str) -> str:
    zfs_datasets = dataset_list()

    for ds in zfs_datasets["datasets"]:
        if exists(ds["mount_path"] + vm_name):
            vm_location = ds["zfs_path"] + "/" + vm_name
            return vm_location
            
        elif ds == len(zfs_datasets["datasets"]) and not exists(ds["mount_path"] + vm_name):
            sys.exit(" 🚦 ERROR: VM location doesn't exist!")


def vm_folder(vm_name:str) -> str:
    zfs_datasets = dataset_list()

    for ds in zfs_datasets["datasets"]:
        if exists(ds["mount_path"] + vm_name):
            return ds["mount_path"] + vm_name
        elif ds == len(zfs_datasets["datasets"]) and not exists(ds["mount_path"] + vm_name):
            sys.exit(" 🚦 ERROR: VM folder doesn't exist!")


def vm_dataset(vm_name:str) -> str:
    zfs_datasets = dataset_list()

    for ds in zfs_datasets["datasets"]:
        if exists("/" + ds["zfs_path"] + "/" + vm_name):
            return ds["zfs_path"]
        elif ds == len(zfs_datasets["datasets"]) and not exists(ds["mount_path"] + vm_name):
            sys.exit(" 🚦 ERROR: VM dataset doesn't exist!")


def list_all_vms() -> list:
    zfs_datasets = dataset_list()

    all_vms = []
    zfs_datasets_list = []
    for ds in zfs_datasets["datasets"]:
        if ds["type"] == "zfs":
            zfs_datasets_list.append(ds["zfs_path"])

    for ds in zfs_datasets_list:
        if exists("/" + ds + "/"):
            dataset_listing = listdir("/" + ds + "/")
            for vm_directory in dataset_listing:
                if exists("/" + ds + "/" + vm_directory + glob_conf_location):
                    all_vms.append(vm_directory)
        else:
            sys.exit(" 🚦 ERROR: Please create 2 zfs datasets: " + zfs_datasets_list)

    if not vmColumnNames:
        print("\n 🚦 ERROR: There are no VMs on this system. To deploy one, use:\n hoster vm deploy\n")
        sys.exit(0)

    all_vms = natsorted(all_vms)
    return all_vms

vm_uptimes = []
def get_vm_uptime(vm_name:str = "zabbix-production") -> dict:
    nonlocal vm_uptimes
    if not vm_uptimes:
        command = "ps axwww -o etimes,command"
        shell_command = subprocess.check_output(command, shell=True)
        vm_uptimes = shell_command.decode("utf-8").split("\n")
        print(vm_uptimes)


def table_output(table_title:bool = False) -> None:
    from rich.console import Console
    from rich.table import Table
    from rich import box

    if not table_title:
        table = Table(box=box.ROUNDED)
    else:
        table = Table(title = "VM List", box=box.ROUNDED)

    table.add_column("Name", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("State", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("CPUs", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("RAM", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Main IP", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("VNC Port", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("VNC Password", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("OS Disk (Used/Overall)", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Uptime", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("OS Comment", justify="center", style="bright_cyan")
    table.add_column("Description", justify="center", style="bright_cyan")


get_vm_uptime()
