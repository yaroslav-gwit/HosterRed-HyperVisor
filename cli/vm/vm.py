# Native Python functions
from os.path import exists
from os import listdir
import subprocess
import random
import json
import time
import sys
import os
import re

# Installed packages/modules
from generate_mac import generate_mac
# from ipaddress import ip_address
from natsort import natsorted
from jinja2 import Template
import psutil
import typer

from rich.console import Console
from rich.table import Table
from rich import box

# Own functions
from cli.host import dataset
from cli.host import host


# SECTION STANDALONE FUNCTIONS
def random_password_generator(capitals: bool = False, numbers: bool = False, length: int = 8, specials: bool = False):
    letters_var = "asdfghjklqwertyuiopzxcvbnm"
    capitals_var = "ASDFGHJKLZXCVBNMQWERTYUIOP"
    numbers_var = "0987654321"
    specials_var = ".,-_!^*?><)(%[]=+$#"

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
    for _i in range(0, length):
        password = password + random.choice(valid_chars_list)

    return password


def mac_address_generator(prefix: str = "58:9C:FC"):
    mac_address = generate_mac.vid_provided(prefix)
    mac_address = mac_address.lower()
    return mac_address


def ip_address_generator(ip_address: str = "10.0.0.0", existing_ip_addresses=None):
    if existing_ip_addresses is None:
        existing_ip_addresses = []
    if len(existing_ip_addresses) < 1:
        existing_ip_addresses = CoreChecks.existing_ip_addresses()

    with open("./configs/networks.json", "r") as file:
        networks_file = file.read()
    networks_dict = json.loads(networks_file)
    networks = networks_dict["networks"][0]

    if ip_address in existing_ip_addresses:
        print("VM with such IP exists: " + ip_address)

    elif ip_address == "10.0.0.0":
        bridge_address = networks["bridge_address"]
        range_start = networks["range_start"]
        range_end = networks["range_end"]

        # Generate full list of IPs for the specified range
        bridge_split = bridge_address.split(".")
        del bridge_split[-1]
        bridge_join = ".".join(bridge_split) + "."

        ip_address_list = []
        for number in range(range_start, range_end + 1):
            ip_address = bridge_join + str(number)
            ip_address_list.append(ip_address)

        ip_address = ip_address_list[0]
        number = range_start
        while ip_address in existing_ip_addresses:
            number = number + 1
            if number > range_end:
                sys.exit("There are no free IPs left!")
            else:
                ip_address = bridge_join + str(number)

    return ip_address


# SECTION CLASS CORE CHECKS
class CoreChecks:
    def __init__(self, vm_name, disk_image_name="disk0.img"):
        if not vm_name:
            print("Please supply a VM name!")
            sys.exit(1)
        self.vm_name = vm_name
        self.zfs_datasets = dataset.DatasetList().datasets
        self.disk_image_name = disk_image_name
        self.vm_config = VmConfigs(vm_name).vm_config_read()

    def vm_is_live(self):
        if exists("/dev/vmm/" + self.vm_name):
            return True
        else:
            return False

    def vm_is_encrypted(self):
        for ds in self.zfs_datasets["datasets"]:
            if exists(ds["mount_path"] + self.vm_name):
                return ds["encrypted"]

    def vm_in_production(self):
        vm_info_dict = self.vm_config
        if vm_info_dict["live_status"] == "Production" or vm_info_dict["live_status"] == "production":
            return True
        else:
            return False

    def disk_exists(self):
        for ds in self.zfs_datasets["datasets"]:
            if exists(ds["mount_path"] + self.vm_name + "/" + self.disk_image_name):
                return True

    def disk_location(self):
        for ds in self.zfs_datasets["datasets"]:
            image_path = ds["mount_path"] + self.vm_name + "/" + self.disk_image_name
            if exists(image_path):
                return image_path

    def vm_location(self):
        for ds in self.zfs_datasets["datasets"]:
            if exists(ds["mount_path"] + self.vm_name):
                vm_location = ds["zfs_path"] + "/" + self.vm_name
                return vm_location
            elif ds == len(self.zfs_datasets["datasets"]) and not exists(ds["mount_path"] + self.vm_name):
                sys.exit("VM doesn't exist!")

    def vm_folder(self):
        for ds in self.zfs_datasets["datasets"]:
            if exists(ds["mount_path"] + self.vm_name):
                vm_folder = ds["mount_path"] + self.vm_name
                return vm_folder
            elif ds == len(self.zfs_datasets["datasets"]) and not exists(ds["mount_path"] + self.vm_name):
                sys.exit("VM doesn't exist!")

    def vm_dataset(self):
        for ds in self.zfs_datasets["datasets"]:
            if exists("/" + ds["zfs_path"] + "/" + self.vm_name):
                vm_dataset = ds["zfs_path"]
                return vm_dataset
            elif ds == len(self.zfs_datasets["datasets"]) and not exists(ds["mount_path"] + self.vm_name):
                sys.exit("VM doesn't exist!")

    def vm_ip_address(self):
        """
        Get VM's IP address
        """
        vm_info_dict = self.vm_config
        vm_ip_address = vm_info_dict["networks"][0]["ip_address"]
        return vm_ip_address

    def vm_vnc_port(self):
        """
        Get VM's VNC port
        """
        vm_info_dict = self.vm_config
        vm_ip_address = vm_info_dict["vnc_port"]
        return vm_ip_address

    # _ VM START PORTION _#
    def vm_network_interfaces(self):
        vm_config = self.vm_config
        vm_network_interfaces = vm_config["networks"]
        return vm_network_interfaces

    def vm_disks(self):
        vm_config = self.vm_config
        vm_disks = vm_config["disks"]
        return vm_disks

    def vm_cpus(self):
        vm_config = self.vm_config
        vm_cpu = {"cpu_sockets": vm_config.get("cpu_sockets", 1), "cpu_cores": vm_config.get("cpu_cores", 2),
                  "memory": vm_config.get("memory", "1G"), "vnc_port": vm_config.get("vnc_port", 5100),
                  "vnc_password": vm_config.get("vnc_password", "NakHkX09a7pgZUQoEJzI"),
                  "loader": vm_config.get("loader", "uefi"), "live_status": vm_config.get("live_status", "testing")}
        return vm_cpu

    def vm_os_type(self):
        vm_config = self.vm_config
        os_type = vm_config.get("os_type", "default_os_type")
        return os_type

    @staticmethod
    def existing_ip_addresses():
        existing_ip_addresses = []
        for _vm in VmList().plainList:
            ip_address = CoreChecks(vm_name=_vm).vm_ip_address()
            existing_ip_addresses.append(ip_address)
        return existing_ip_addresses


class VmConfigs:
    def __init__(self, vm_name):
        self.vm_name = vm_name
        self.zfs_datasets = dataset.DatasetList().datasets
        self.vm_config = "/vm_config.json"

    def vm_config_read(self):
        for ds in self.zfs_datasets["datasets"]:
            vm_config = ds["mount_path"] + self.vm_name + self.vm_config
            if exists(vm_config):
                with open(vm_config, 'r') as file:
                    vm_info_raw = file.read()
                vm_info_dict = json.loads(vm_info_raw)
                return vm_info_dict
            elif ds == self.zfs_datasets["datasets"][-1] and not exists(vm_config):
                print("Sorry, config file was not found for " + self.vm_name + " path: " + vm_config)
                sys.exit(1)

    def vm_config_manual_edit(self):
        text_editor = os.getenv("EDITOR")
        check_editor_output = subprocess.check_output(f"which {text_editor}", stderr=subprocess.DEVNULL, text=True,
                                                      shell=True)
        if len(check_editor_output) < 1:
            text_editor = "nano"
        for ds in self.zfs_datasets["datasets"]:
            vm_config = ds["mount_path"] + self.vm_name + self.vm_config
            if exists(vm_config):
                command = f"{text_editor} {vm_config}"
                subprocess.run(command, shell=True)
                return
            elif ds == self.zfs_datasets["datasets"][-1] and not exists(vm_config):
                print("Sorry, config file was not found for " + self.vm_name + " path: " + vm_config)
                sys.exit(1)

    @staticmethod
    def vm_config_write():
        print("This function will write config files to the required directories")


class VmList:
    def __init__(self, return_value: bool = False):
        self.zfs_datasets = dataset.DatasetList().datasets
        self.uptime_command_output = ""
        self.return_value = return_value

        vm_column_names = []
        zfs_datasets_list = []
        for ds in self.zfs_datasets["datasets"]:
            if ds["type"] == "zfs":
                zfs_datasets_list.append(ds["zfs_path"])

        for ds in zfs_datasets_list:
            if exists("/" + ds + "/"):
                _dataset_listing = listdir("/" + ds + "/")
                for vm_directory in _dataset_listing:
                    if exists("/" + ds + "/" + vm_directory + "/vm_config.json"):
                        vm_column_names.append(vm_directory)
            else:
                sys.exit(" 🚦 ERROR: Please create 2 zfs datasets: " + str(zfs_datasets_list))

        if not vm_column_names and not self.return_value:
            print("\n 🚦 ERROR: There are no VMs on this system. To deploy one, use:\n hoster vm deploy\n")
            sys.exit(0)

        self.plainList = vm_column_names.copy()
        self.vm_natsorted_list = natsorted(vm_column_names)

    def table_output(self, table_title: bool = False):
        vm_list = self.vm_natsorted_list

        if len(vm_list) < 1:
            print("\n 🚦 ERROR: There are no VMs on this system. To deploy one, use:\n hoster vm deploy\n")
            sys.exit(0)

        # GENERATE AND PRINT THE TABLE
        if not table_title:
            table = Table(box=box.ROUNDED, show_lines=True, )
        else:
            table = Table(title=" VM List", box=box.ROUNDED, show_lines=True, title_justify="left")
        table.add_column("#", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Name", justify="left", style="bright_cyan", no_wrap=True)
        table.add_column("State", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("CPUs", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("RAM", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Main IP", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("VNC\nPort", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("VNC\nPassword", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("OS Disk\n(Used/Total)", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("OS Comment", justify="left", style="bright_cyan", no_wrap=True)
        table.add_column("Uptime", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Description", justify="center", style="bright_cyan", no_wrap=True)

        vm_index = 0
        for vm_name in vm_list:
            # GET SPECIFIC VM CONFIG
            vm_config = VmConfigs(vm_name).vm_config_read()

            # VM LIVE CHECK
            vm_is_live = CoreChecks(vm_name).vm_is_live()
            if vm_is_live:
                state = "🟢"
            elif vm_config.get("parent_host") != host.get_hostname():
                state = "💾"
            else:
                state = "🔴"
            if CoreChecks(vm_name).vm_is_encrypted():
                state = state + "🔒"
            if CoreChecks(vm_name).vm_in_production():
                state = state + "🔁"
            vm_state = state

            # GET CPU, RAM, VNC INFO, OS TYPE
            vm_cpu = vm_config.get("cpu_cores", "-")
            vm_ram = vm_config.get("memory", "-")
            vm_vnc_port = vm_config.get("vnc_port", "-")
            vm_vnc_password = vm_config.get("vnc_password", "-")
            vm_os_type = vm_config.get("os_comment", "-")

            # GET DISK INFO
            vm_disks = vm_config.get("disks", "-")
            disk_image_name = vm_disks[0].get("disk_image", "-")
            if CoreChecks(vm_name, disk_image_name).disk_exists():
                image_path = CoreChecks(vm_name, disk_image_name).disk_location()
                command_size = "ls -ahl " + image_path + " | awk '{ print $5 }'"
                command_used = "du -h " + image_path + " | awk '{ print $1 }'"
                shell_command_size = subprocess.check_output(command_size, shell=True)
                shell_command_used = subprocess.check_output(command_used, shell=True)
                disk_size = shell_command_size.decode("utf-8").split()[0]
                disk_used = shell_command_used.decode("utf-8").split()[0]
                vm_disk_final_output = disk_used + "/" + disk_size
            else:
                vm_disk_final_output = "-"

            # GET IP INFO
            vm_networks = vm_config.get("networks", "-")
            vm_ip_addr = vm_networks[0].get("ip_address", "-")
            # vmColumnOsType = ["Debian 10" if var == "debian10" else var for var in vmColumnOsType]

            # GET VM UPTIME
            re_match_bhyve = re.compile(".*bhyve: " + vm_name + ".*")
            vm_uptime = ""
            if vm_is_live and len(self.uptime_command_output) < 1:
                command = "ps axwww -o etimes,command"
                shell_command = subprocess.check_output(command, shell=True)
                self.uptime_command_output = shell_command

                vm_uptimes = shell_command.decode("utf-8").split("\n")
                for i in vm_uptimes:
                    if re_match_bhyve.match(i):
                        vm_uptime = i.split()[0]
                        vm_uptime = host.human_readable_uptime(int(vm_uptime))

            elif vm_is_live and len(self.uptime_command_output) > 1:
                shell_command = self.uptime_command_output
                vm_uptimes = shell_command.decode("utf-8").split("\n")
                for i in vm_uptimes:
                    if re_match_bhyve.match(i):
                        vm_uptime = i.split()[0]
                        vm_uptime = host.human_readable_uptime(int(vm_uptime))
            else:
                vm_uptime = "-"

            # GET VM DESCRIPTION
            if vm_config.get("parent_host") != host.get_hostname():
                vm_description = vm_config.get("parent_host", "-")
                vm_description = ("💾⏩ " + vm_description)
            else:
                vm_description = vm_config.get("description", "-")

            vm_index = vm_index + 1
            table.add_row(
                str(vm_index),
                vm_name,
                vm_state,
                vm_cpu,
                vm_ram,
                vm_ip_addr,
                vm_vnc_port,
                vm_vnc_password,
                vm_disk_final_output,
                vm_os_type,
                vm_uptime,
                vm_description,
            )

        Console().print(table)

    def json_output(self):
        vm_list_dict = self.vm_natsorted_list
        vm_list_json = json.dumps(vm_list_dict, indent=2)

        return vm_list_json


# SECTION CLASS VM DEPLOY
class VmDeploy:
    def __init__(self, vm_name: str = "test-vm", ip_address: str = "10.0.0.0", os_type: str = "debian11",
                 vnc_port: int = 5900, dataset_id: int = 0):
        # _ Load networks config _#
        with open("./configs/networks.json", "r") as file:
            networks_file = file.read()
        networks_dict = json.loads(networks_file)
        self.networks = networks_dict["networks"][0]

        # _ Load host config _#
        with open("./configs/host.json", "r") as file:
            host_file = file.read()
        host_dict = json.loads(host_file)
        self.host_dict = host_dict

        self.vm_name = vm_name
        self.ip_address = ip_address

        self.existing_ip_addresses = []
        for _vm in VmList(return_value=True).plainList:
            ip_address = CoreChecks(vm_name=_vm).vm_ip_address()
            self.existing_ip_addresses.append(ip_address)

        self.existing_vms = VmList(return_value=True).plainList

        # OS Type Settings
        os_type_list = ["debian11", "ubuntu2004", "freebsd13ufs", "freebsd13zfs", "almalinux8", "rockylinux8"]
        if os_type in os_type_list:
            self.os_type = os_type
        else:
            os_type_list = " ".join(os_type_list)
            sys.exit(" 🚦 ERROR: Sorry this OS is not supported. Here is the list of supported OSes:\n" + os_type_list)

        self.vnc_port = vnc_port

        if vm_name == "test-vm":
            self.live_status = "testing"
        else:
            self.live_status = "production"

        self.dataset_id = dataset_id

    @staticmethod
    def vm_vnc_port_generator(vnc_port: int = 5900):
        existing_vnc_ports = []
        allowed_vnc_ports = []
        for _port in range(5900, 6100):
            allowed_vnc_ports.append(_port)

        for _vm in VmList(return_value=True).plainList:
            _vm_vnc_port = CoreChecks(vm_name=_vm).vm_vnc_port()
            _vm_vnc_port = int(_vm_vnc_port)
            existing_vnc_ports.append(_vm_vnc_port)

        if vnc_port not in allowed_vnc_ports:
            sys.exit("You can't assign this port to your VM! Allowed range: 5900-6100")

        while vnc_port in existing_vnc_ports:
            if 5900 <= vnc_port <= 6100:
                vnc_port = vnc_port + 1
            else:
                sys.exit("We ran out of available VNC ports!")

        return vnc_port

    @staticmethod
    def vm_name_generator(vm_name: str, existing_vms):
        # Generate test VM name and number
        number = 1
        if vm_name in existing_vms:
            print("VM with this name exists: " + vm_name)
            sys.exit(0)
        elif vm_name == "test-vm":
            vm_name = "test-vm-" + str(number)
            while vm_name in existing_vms:
                number = number + 1
                vm_name = "test-vm-" + str(number)
        else:
            vm_name = vm_name
        return vm_name

    @staticmethod
    def ip_address_generator(ip_address: str, networks, existing_ip_addresses, vm_name: str):
        if ip_address in existing_ip_addresses and vm_name != "test-vm":
            print("VM with such IP exists: " + vm_name + "/" + ip_address)

        elif ip_address == "10.0.0.0":
            bridge_address = networks["bridge_address"]
            range_start = networks["range_start"]
            range_end = networks["range_end"]

            # Generate full list of IPs for the specified range
            bridge_split = bridge_address.split(".")
            del bridge_split[-1]
            bridge_join = ".".join(bridge_split) + "."

            ip_address_list = []
            for number in range(range_start, range_end + 1):
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

    @staticmethod
    def random_password_generator(capitals: bool = False, numbers: bool = False, length: int = 8,
                                  specials: bool = False):
        letters_var = "asdfghjklqwertyuiopzxcvbnm"
        capitals_var = "ASDFGHJKLZXCVBNMQWERTYUIOP"
        numbers_var = "0987654321"
        specials_var = ".,-_!^*?><)(%[]=+$#"

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
        for _i in range(0, length):
            password = password + random.choice(valid_chars_list)

        return password

    @staticmethod
    def mac_address_generator(prefix: str = "58:9C:FC"):
        mac_address = generate_mac.vid_provided(prefix)
        mac_address = mac_address.lower()
        return mac_address

    def dns_registry(self):
        dns_registry = {}
        vms_and_ips = []

        dns_registry["host_dns_acls"] = self.host_dict["host_dns_acls"]
        dns_registry["vms_and_ips"] = vms_and_ips

        for vm_index, vm_name in enumerate(self.existing_vms):
            vm_and_ip_dict = {}
            ip_address = CoreChecks(self.existing_vms[vm_index]).vm_ip_address()
            vm_and_ip_dict["vm_name"] = vm_name
            vm_and_ip_dict["ip_address"] = ip_address
            vms_and_ips.append(vm_and_ip_dict)

        # Read Unbound template
        with open("./templates/unbound.conf", "r") as file:
            template = file.read()
        # Render Unbound template
        template = Template(template)
        template = template.render(dns_registry=dns_registry)
        # Write Unbould template
        with open("/var/unbound/unbound.conf", "w") as file:
            file.write(template)
        # Reload the Unbound service
        command = "service local_unbound reload"
        subprocess.run(command, shell=True, stdout=subprocess.DEVNULL)

        return

    def deploy(self):
        output_dict = {}
        output_dict["vm_name"] = VmDeploy.vm_name_generator(vm_name=self.vm_name, existing_vms=self.existing_vms)
        output_dict["ip_address"] = VmDeploy.ip_address_generator(ip_address=self.ip_address, networks=self.networks,
                                                                  existing_ip_addresses=self.existing_ip_addresses,
                                                                  vm_name=self.vm_name)
        output_dict["os_type"] = self.os_type
        output_dict["root_password"] = VmDeploy.random_password_generator(length=41, capitals=True, numbers=True)
        output_dict["user_password"] = VmDeploy.random_password_generator(length=41, capitals=True, numbers=True)
        output_dict["vnc_port"] = VmDeploy.vm_vnc_port_generator(vnc_port=self.vnc_port)
        output_dict["vnc_password"] = VmDeploy.random_password_generator(length=8, capitals=True, numbers=True)
        output_dict["mac_address"] = VmDeploy.mac_address_generator()
        network0 = self.networks
        network_bridge_name = network0["bridge_name"]
        network_bridge_address = network0["bridge_address"]
        output_dict["network_bridge_name"] = network_bridge_name
        output_dict["network_bridge_address"] = network_bridge_address
        output_dict["live_status"] = self.live_status
        output_dict["host"] = host.get_hostname()

        if self.os_type == "ubuntu2004":
            output_dict["os_comment"] = "Ubuntu 20.04"
        elif self.os_type == "debian11":
            output_dict["os_comment"] = "Debian 11"
        elif self.os_type == "freebsd13ufs":
            output_dict["os_comment"] = "FreeBSD 13 UFS"
        elif self.os_type == "freebsd13zfs":
            output_dict["os_comment"] = "FreeBSD 13 ZFS"
        elif self.os_type == "almalinux8":
            output_dict["os_comment"] = "AlmaLinux 8"
        elif self.os_type == "rockylinux8":
            output_dict["os_comment"] = "RockyLinux 8"
        else:
            output_dict["os_comment"] = self.os_type

        # Cloud Init Section
        host_dict = self.host_dict
        vm_ssh_keys = []
        for _key in host_dict["host_ssh_keys"]:
            _ssh_key = _key["key_value"]
            vm_ssh_keys.append(_ssh_key)
        output_dict["random_instanse_id"] = VmDeploy.random_password_generator(length=5)
        output_dict["vm_ssh_keys"] = vm_ssh_keys

        dataset_id = self.dataset_id
        working_dataset = dataset.DatasetList().datasets["datasets"][dataset_id]["zfs_path"]
        working_dataset_path = dataset.DatasetList().datasets["datasets"][dataset_id]["mount_path"]

        # Clone a template using ZFS clone
        template_ds = working_dataset + "/template-" + output_dict["os_type"]
        template_folder = working_dataset_path + "template-" + output_dict["os_type"]
        if exists(template_folder):
            snapshot_name = "@deployment_" + output_dict["vm_name"] + "_" + VmDeploy.random_password_generator(length=7,
                                                                                                               numbers=True)
            command = "zfs snapshot " + template_ds + snapshot_name
            # print(command)
            subprocess.run(command, shell=True)

            command = "zfs clone " + template_ds + snapshot_name + " " + working_dataset + "/" + output_dict["vm_name"]
            # print(command)
            subprocess.run(command, shell=True)
        else:
            sys.exit(" ⛔ FATAL! Template specified doesn't exist: " + template_folder)

        new_vm_folder = working_dataset_path + output_dict["vm_name"] + "/"
        if exists(new_vm_folder):
            # Read VM template
            with open("./templates/vm_config_template.json", "r") as file:
                template = file.read()
            # Render VM template
            template = Template(template)
            template = template.render(output_dict=output_dict)
            # Write VM template
            with open(new_vm_folder + "vm_config.json", "w") as file:
                file.write(template)

            CloudInit(vm_name=output_dict["vm_name"], vm_folder=new_vm_folder, vm_ssh_keys=vm_ssh_keys,
                      os_type=output_dict["os_type"], ip_address=output_dict["ip_address"],
                      network_bridge_address=output_dict["network_bridge_address"],
                      root_password=output_dict["root_password"],
                      user_password=output_dict["user_password"], mac_address=output_dict["mac_address"]).deploy()

        else:
            sys.exit(" ⛔ FATAL! Template specified doesn't exist: " + template_folder)

        return {"status": "success", "vm_name": output_dict["vm_name"]}


class Operation:
    @staticmethod
    def snapshot(vm_name: str, stype: str = "custom", keep: int = 3) -> None:
        """ Function responsible for taking VM Snapshots """

        snapshot_type = stype
        snapshots_to_keep = keep
        snapshot_type_list = ["replication", "custom", "hourly", "daily", "weekly", "monthly", "yearly"]
        if vm_name in VmList().plainList:
            date_now = time.strftime("%Y-%m-%d_%H-%M-%S", time.localtime())
            snapshot_name = snapshot_type + "_" + date_now
            command = "zfs snapshot " + CoreChecks(vm_name).vm_location() + "@" + snapshot_name
            subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)
            # DEBUG
            print(" 🔷 DEBUG: New snapshot was taken: " + command)
        else:
            sys.exit(" 🚫 CRITICAL: Can't snapshot! VM " + vm_name + " doesn't exist on this system!")

        # Remove old snapshots
        if snapshot_type != "custom":
            # Get the snapshot list
            command = "zfs list -r -t snapshot " + CoreChecks(
                vm_name).vm_location() + " | tail +2 | awk '{ print $1 }' | grep " + snapshot_type
            shell_command = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT)
            vm_zfs_snapshot_list = shell_command.decode("utf-8").split()

            # Generate list of snapshots to delete
            vm_zfs_snapshots_to_delete = vm_zfs_snapshot_list.copy()
            if len(vm_zfs_snapshots_to_delete) > 0 and len(vm_zfs_snapshots_to_delete) > snapshots_to_keep:
                for zfs_snapshot in range(0, snapshots_to_keep):
                    del vm_zfs_snapshots_to_delete[-1]
                # Remove the old snapshots
                for vm_zfs_snapshot_to_delete in vm_zfs_snapshots_to_delete:
                    command = "zfs destroy " + vm_zfs_snapshot_to_delete
                    subprocess.run(command, shell=True)
                    print(" 🔷 DEBUG: Old snapshot was removed: " + command)
            else:
                print(" 🔷 DEBUG: VM " + vm_name + " doesn't have any '" + snapshot_type + "' snapshots to delete")

    @staticmethod
    def destroy(vm_name: str, force: bool = False):
        """
        Function responsible for completely removing VMs from the system
        """
        if force == True and CoreChecks(vm_name).vm_is_live():
            kill(vm_name=vm_name)
            time.sleep(3)

        if vm_name not in VmList().plainList:
            print(" 🔶 INFO: VM doesn't exist on this system.")
        elif CoreChecks(vm_name).vm_is_live():
            print(" 🔴 WARNING: VM is still running, you'll have to stop (or kill) it first: " + vm_name)
        else:
            command = "zfs destroy -rR " + CoreChecks(vm_name).vm_location()
            # ADD DEBUG/FAKE RUN
            shell_command = subprocess.check_output(command, shell=True)
            print(" 🔶 INFO: The VM was destroyed: " + command)

    @staticmethod
    def kill(vm_name: str, quiet: bool = False):
        """
        Function that forcefully kills the VM
        """
        if vm_name not in VmList().plainList:
            sys.exit("VM doesn't exist on this system.")
        elif CoreChecks(vm_name).vm_is_live():
            # This code block is a duplicate. Another one exists in stop section.
            command = "ps axf | grep -v grep | grep 'nmdm-" + vm_name + "' | awk '{ print $1 }'"
            shell_command = subprocess.check_output(command, shell=True)
            console_list = shell_command.decode("utf-8").split()
            for _console in console_list:
                if _console:
                    command = "kill -SIGKILL " + _console
                    subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

            # Find and kill the VM process
            try:
                pid_file = "/var/run/" + vm_name + ".pid"
                if exists(pid_file):
                    command = "cat /var/run/" + vm_name + ".pid"
                    shell_command = subprocess.check_output(command, shell=True)
                    parent_pid = int(shell_command.decode("utf-8").split()[0])
                    child_pid = psutil.Process(parent_pid).children()[-1].pid
                    running_vm_pid = str(child_pid)
                    command = "kill -s SIGKILL " + running_vm_pid
                    shell_command = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT)

                    if not quiet:
                        print(" 🔶 INFO: Could not find the PID file for: " + vm_name)
                    command = "top -b -d1 -a all | grep \"/" + vm_name + "/\" | grep bash | awk '{print $1}'"
                    shell_command = subprocess.check_output(command, shell=True)
                    console_list = shell_command.decode("utf-8").split()
                    command = "kill -s SIGKILL " + console_list[0]
                    subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                    if not quiet:
                        print(" 🔶 INFO: Forcefully killed the VM process: " + console_list[0] + " " + vm_name)


                else:
                    if not quiet:
                        print(" 🔶 INFO: Could not find the PID file for: " + vm_name)
                    command = "top -b -d1 -a all | grep \"/" + vm_name + "/\" | grep bash | awk '{print $1}'"
                    shell_command = subprocess.check_output(command, shell=True)
                    console_list = shell_command.decode("utf-8").split()
                    command = "kill -s SIGKILL " + console_list[0]
                    subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                    if not quiet:
                        print(" 🔶 INFO: Forcefully killed the VM process: " + console_list[0] + " " + vm_name)

            except Exception as e:
                print(" 🔶 INFO: Could not find the PID file for: " + vm_name)
                if not quiet:
                    print(" 🔶 INFO: Could not find the PID file for: " + vm_name)
                command = "top -b -d1 -a all | grep \"/" + vm_name + "/\" | grep bash | awk '{print $1}'"
                shell_command = subprocess.check_output(command, shell=True)
                console_list = shell_command.decode("utf-8").split()
                command = "kill -s SIGKILL " + console_list[0]
                subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                if not quiet:
                    print(" 🔶 INFO: Forcefully killed the VM process: " + console_list[0] + " " + vm_name)
                # print(e)

            # command = "ps axf | grep -v grep | grep " + vm_name + " | grep bhyve: | awk '{ print $1 }'"
            # shell_command = subprocess.check_output(command, shell=True)
            # try:
            #     running_vm_pid = shell_command.decode("utf-8").split()[0]
            #     command = "kill -SIGKILL " + running_vm_pid
            #     subprocess.run(command, shell=True)
            # except:
            #     print(" 🔶 INFO: Could not find the process for the VM: " + vm_name)

            # This block is a duplicate. Creating a function would be a good idea for the future!
            command = "ifconfig | grep " + vm_name + " | awk '{ print $2 }'"
            shell_command = subprocess.check_output(command, shell=True)
            tap_interface_list = shell_command.decode("utf-8").split()

            command = "bhyvectl --destroy --vm=" + vm_name
            subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

            time.sleep(1)

            if tap_interface_list:
                for tap in tap_interface_list:
                    if tap:
                        command = "ifconfig " + tap + " destroy"
                        subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            if not quiet:
                print(" 🔶 INFO: Killed the VM: " + vm_name)
        else:
            # This block is a duplicate. Creating a function would be a good idea for the future!
            command = "ifconfig | grep " + vm_name + " | awk '{ print $2 }'"
            shell_command = subprocess.check_output(command, shell=True)
            tap_interface_list = shell_command.decode("utf-8").split()
            if tap_interface_list:
                for tap in tap_interface_list:
                    if tap:
                        command = "ifconfig " + tap + " destroy"
                        subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            if not quiet:
                print(" 🔶 INFO: VM is already dead: " + vm_name + "!")

        # Remove PID file if it still exists
        try:
            command = "rm /var/run/" + vm_name + ".pid"
            subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        except:
            pass

    @staticmethod
    def start(vm_name: str) -> None:
        vm_config = VmConfigs(vm_name).vm_config_read()
        if vm_config["parent_host"] != host.get_hostname():
            print(" 🚦 ERROR: VM is a backup from another host: " + vm_config[
                "parent_host"] + ". Run 'hoster vm cireset " + vm_name + "' if you want to use it on this host!")
            return
        elif CoreChecks(vm_name).vm_is_live():
            print(" 🔶 INFO: VM is already live: " + vm_name)
        elif vm_name in VmList().plainList:
            print(" 🔶 INFO: Starting the VM: " + vm_name)

            # _ NETWORKING - Create required TAP interfaces _#
            vm_network_interfaces = CoreChecks(vm_name).vm_network_interfaces()
            tap_interface_number = 0
            tap_interface_list = []

            for interface in range(len(vm_network_interfaces)):
                command = "ifconfig | grep -G '^tap' | awk '{ print $1 }' | sed s/://"
                shell_command = subprocess.check_output(command, shell=True)
                existing_tap_interfaces = shell_command.decode("utf-8").split()
                tap_interface = "tap" + str(tap_interface_number)
                while tap_interface in existing_tap_interfaces:
                    tap_interface_number = tap_interface_number + 1
                    tap_interface = "tap" + str(tap_interface_number)
                # print(tap_interface)

                command = "ifconfig " + tap_interface + " create"
                # print(command)
                subprocess.run(command, shell=True)

                command = "ifconfig vm-" + vm_network_interfaces[interface]["network_bridge"] + " addm " + tap_interface
                # print(command)
                subprocess.run(command, shell=True)

                command = "ifconfig vm-" + vm_network_interfaces[interface]["network_bridge"] + " up"
                # print(command)
                subprocess.run(command, shell=True)

                command = 'ifconfig ' + tap_interface + ' description ' + '"' + tap_interface + ' ' + vm_name + ' ' + 'interface' + str(
                    interface) + '"'
                # print(command)
                subprocess.run(command, shell=True)

                tap_interface_list.append(tap_interface)

            # _ NEXT SECTION _#
            command1 = "bhyve -HAw -s 0:0,hostbridge -s 31,lpc "

            bhyve_pci_1 = 2
            bhyve_pci_2 = 0
            space = " "
            if len(vm_network_interfaces) > 1:
                for interface in range(len(vm_network_interfaces)):
                    network_adaptor_type = vm_network_interfaces[interface]["network_adaptor_type"]
                    generic_network_text = "," + network_adaptor_type + ","
                    if interface == 0:
                        network_final = "-s " + str(bhyve_pci_1) + ":" + str(bhyve_pci_2) + generic_network_text + \
                                        tap_interface_list[interface] + ",mac=" + vm_network_interfaces[interface][
                                            "network_mac"]
                    else:
                        bhyve_pci_2 = bhyve_pci_2 + 1
                        network_final = network_final + space + "-s " + str(bhyve_pci_1) + ":" + str(
                            bhyve_pci_2) + generic_network_text + tap_interface_list[interface] + ",mac=" + \
                                        vm_network_interfaces[interface]["network_mac"]
            else:
                network_adaptor_type = vm_network_interfaces[0]["network_adaptor_type"]
                generic_network_text = "," + network_adaptor_type + ","
                network_final = "-s " + str(bhyve_pci_1) + ":" + str(bhyve_pci_2) + generic_network_text + \
                                tap_interface_list[0] + ",mac=" + vm_network_interfaces[0]["network_mac"]

            command2 = network_final

            bhyve_pci = 3
            vm_disks = CoreChecks(vm_name).vm_disks()
            if len(vm_disks) > 1:
                for disk in range(len(vm_disks)):
                    generic_disk_text = ":0," + vm_disks[disk]["disk_type"] + ","
                    disk_image = vm_disks[disk]["disk_image"]
                    if disk == 0:
                        disk_final = " -s " + str(bhyve_pci) + generic_disk_text + CoreChecks(vm_name=vm_name,
                                                                                              disk_image_name=disk_image).disk_location()
                    else:
                        bhyve_pci = bhyve_pci + 1
                        disk_final = disk_final + " -s " + str(bhyve_pci) + generic_disk_text + CoreChecks(
                            vm_name=vm_name, disk_image_name=disk_image).disk_location()
            else:
                generic_disk_text = ":0," + vm_disks[0]["disk_type"] + ","
                disk_image = vm_disks[0]["disk_image"]
                disk_final = " -s " + str(bhyve_pci) + generic_disk_text + CoreChecks(vm_name=vm_name,
                                                                                      disk_image_name=disk_image).disk_location()

            command3 = disk_final

            os_type = CoreChecks(vm_name).vm_os_type()
            vm_cpus = CoreChecks(vm_name).vm_cpus()
            command5 = " -c sockets=" + vm_cpus["cpu_sockets"] + ",cores=" + vm_cpus["cpu_cores"] + " -m " + vm_cpus[
                "memory"]

            bhyve_pci = bhyve_pci + 1
            vnc_port = str(vm_cpus["vnc_port"])
            vnc_password = vm_cpus["vnc_password"]
            command6 = " -s " + str(bhyve_pci) + ":" + str(
                bhyve_pci_2) + ",fbuf,tcp=0.0.0.0:" + vnc_port + ",w=1280,h=1024,password=" + vnc_password

            bhyve_pci = bhyve_pci + 1
            if vm_cpus["loader"] == "bios":
                command7 = " -s " + str(bhyve_pci) + ":" + str(
                    bhyve_pci_2) + ",xhci,tablet -l com1,/dev/nmdm-" + vm_name + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI_CSM.fd -u " + vm_name
                # command = command1 + command2 + command3 + command4 + command5 + command6 + command7
                command = command1 + command2 + command3 + command5 + command6 + command7
            elif vm_cpus["loader"] == "uefi":
                command7 = " -s " + str(
                    bhyve_pci) + ",xhci,tablet -l com1,/dev/nmdm-" + vm_name + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI.fd -u " + vm_name
                # command = command1 + command2 + command3 + command4 + command5 + command6 + command7
                command = command1 + command2 + command3 + command5 + command6 + command7
            else:
                print(" 🚦 ERROR: Loader is not supported!")

            vm_folder = CoreChecks(vm_name).vm_folder()

            command = "nohup ./cli/shell_helpers/vm_start.sh " + '"' + command + '"' + " " + vm_name + " >> " + vm_folder + "/vm.log 2>&1 &"
            subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT)

            # GENERATE VM SERVICE FILE FOR SUPERVISORD
            # if CoreChecks(vm_name).vm_in_production:
            # vm_autostart = "true"
            # else:
            # vm_autostart = "false"

            # with open("./configs/service.vm.conf.jinja", "r") as file:
            #     vm_service_template = file.read()
            # vm_service_template = Template(vm_service_template)
            # vm_service_template = vm_service_template.render(
            #     vm_name=vm_name,
            #     command=command,
            #     # autostart=vm_autostart,
            #     vm_folder=vm_folder,
            # )
            # with open("/var/run/" + vm_name + ".vm.conf", "w") as file:
            #     file.write(vm_service_template)

            # print(vm_service_template)

            # command = "supervisorctl -u user -p 123 update " + vm_name
            # print(command)
            # command = "supervisorctl -u user -p 123 start " + vm_name
            # print(command)
            # subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT)

            # _EOF_ GENERATE VM SERVICE FILE FOR SUPERVISORD

        else:
            print(" 🚦 ERROR: Such VM '" + vm_name + "' doesn't exist!")

    @staticmethod
    def stop(vm_name: str) -> None:

        """ Gracefully stop the VM """

        if vm_name not in VmList().plainList:
            print(" 🚦 ERROR: VM doesn't exist on this system.")
        elif CoreChecks(vm_name).vm_is_live():
            print(" 🔶 INFO: Gracefully stopping the VM: " + vm_name)

            # Send the shutdown signal to the VM process
            vm_process_list = []
            command = "pgrep -lf \"bhyve:\" || true"
            shell_command = subprocess.check_output(command, shell=True, text=True, stderr=subprocess.DEVNULL)
            vm_process = shell_command.split("\n")

            re_match_bhyve_process = re.compile(".*" + vm_name)
            for i in vm_process:
                if re_match_bhyve_process.match(i):
                    vm_process_list.append(i.split()[0])

            for process in vm_process_list:
                if process:
                    command = "kill -s TERM " + process
                    print(" 🔶 INFO: Sending TERM signal to the Bhyve VM Process: " + command)
                    subprocess.run(command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

            print(" 🟢 SUCCESS: The VM is fully stopped now: " + vm_name)

        else:
            print(" 🔶 INFO: VM is already stopped: " + vm_name)

    @staticmethod
    def show_log(vm_name: str) -> None:
        """ Show the live VM's log output """
        log_file_location = CoreChecks(vm_name).vm_folder() + "vm.log"
        if vm_name not in VmList().plainList:
            print(" 🚦 ERROR: VM doesn't exist on this system.")
        elif exists(log_file_location):
            subprocess.run("tail -f " + log_file_location, shell=True)
        else:
            print(" 🔶 INFO: Could not find a log file for: " + vm_name)


class ZFSReplication:
    # typer replicate command

    @staticmethod
    def pull(self) -> None:
        print(" 🚫 FATAL: Sorry this function has not been implemented yet!")
        sys.exit(0)

    @staticmethod
    def push(vm_name: str, ep_address: str, ep_port: str = "22") -> None:
        if vm_name not in VmList().plainList:
            sys.exit(" 🚦 ERROR: This VM doesn't exist: " + vm_name)

        # CHECK IF THE SSH CONNECTION IS AVAILABLE
        command = "timeout 2 ssh root@" + ep_address + " -p" + ep_port + " echo 1 || echo 2"
        shell_output = subprocess.check_output(command, shell=True).decode("UTF-8").split()[0]
        if shell_output == "2":
            print(" 🚫 FATAL: Sorry can't reach the endpoint specified!")
            sys.exit(1)

        # CHECK IF THE ENDPOINT SYSTEM IS A FREEBSD SYSTEM
        command = "timeout 2 ssh root@" + ep_address + " -p" + ep_port + " uname || echo 2"
        shell_output = subprocess.check_output(command, shell=True).decode("UTF-8").split()[0]
        if shell_output == "2" or (shell_output != "FreeBSD"):
            print(" 🚫 FATAL: Sorry the endpoint is not a FreeBSD system!")
            sys.exit(1)

        # Check if VM is from this host:
        vm_config_dict = VmConfigs(vm_name).vm_config_read()
        host_name = host.get_hostname()
        if vm_config_dict["parent_host"] != host_name:
            sys.exit(" 🚦 ERROR: VM is already a backup from another host, can't replicate: " + vm_name)

        # Make a replication snapshot
        Operation.snapshot(vm_name=vm_name, stype="replication")

        vm_dataset = CoreChecks(vm_name).vm_dataset() + "/" + vm_name
        print(" 🟢 INFO: Dataset we are working with: " + vm_dataset)

        command = "zfs list -r -t snapshot " + vm_dataset + " | tail +2 | awk '{ print $1 }'"
        shell_command = subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT)
        vm_zfs_snapshot_list = shell_command.decode("utf-8").split("\n")
        for item in vm_zfs_snapshot_list:
            if not item:
                vm_zfs_snapshot_list.remove(item)
            elif item == "no datasets available":
                vm_zfs_snapshot_list.remove(item)

        """ In case of future debugging """
        # print("List of local snapshots:")
        # for item in vm_zfs_snapshot_list:
        #     print(item)

        # Remote snapshot list
        command = 'echo "if [[ -d /' + vm_dataset + ' ]]; then zfs list -r -t snapshot ' + vm_dataset + '; fi" | ssh ' + ep_address + ' /usr/local/bin/bash | tail +2 | ' + "awk '{ print $1 }'"
        shell_command = subprocess.check_output(command, shell=True)
        remote_zfs_snapshot_list = shell_command.decode("utf-8").split()

        for item in remote_zfs_snapshot_list:
            if not item:
                remote_zfs_snapshot_list.remove(item)
            elif item == "no datasets available":
                remote_zfs_snapshot_list.remove(item)

        """ In case of future debugging """
        # print("List of remote snapshots:")
        # for item in remote_zfs_snapshot_list:
        # print(item)

        if vm_zfs_snapshot_list == remote_zfs_snapshot_list:
            print(" 🔷 DEBUG: The backup system is already up to date.")
            sys.exit(0)

        # Revert to a last snapshot to avoid dealing with differences
        replication_snapshot_list = []
        replication_snapshot_list.extend(remote_zfs_snapshot_list)
        if len(replication_snapshot_list) >= 1:
            for loop_item in replication_snapshot_list:
                if not re.match(".*@replication.*", loop_item):
                    replication_snapshot_list.remove(loop_item)
        if len(replication_snapshot_list) >= 1:
            command = "ssh " + ep_address + " zfs rollback -r " + replication_snapshot_list[-1]
            print(" 🔷 DEBUG: Reverting back to the latest replication snapshot: " + command)
            subprocess.run(command, shell=True)
            # for line in subprocess.check_output(command, shell=True, stderr=subprocess.STDOUT):
            # print(line.split("Line from Python: \n"))

        # Difference list
        to_delete_snapshot_list = []
        to_delete_snapshot_list.extend(remote_zfs_snapshot_list)
        for zfs_snapshot in vm_zfs_snapshot_list:
            if zfs_snapshot in to_delete_snapshot_list:
                to_delete_snapshot_list.remove(zfs_snapshot)
        """ In case of future debugging """
        # print("To delete list: ")
        # for item in to_delete_snapshot_list:
        # print(item)

        if len(to_delete_snapshot_list) != 0:
            # print("Removing old snapshots from the remote system:")
            for item in to_delete_snapshot_list:
                command = "ssh " + ep_address + " zfs destroy " + item
                subprocess.run(command, shell=True)

        # Generate a lists of snapshots to transfer
        for rsnapshot_index, rsnapshot_value in enumerate(remote_zfs_snapshot_list):
            if rsnapshot_index != len(remote_zfs_snapshot_list) - 1:
                if rsnapshot_value in vm_zfs_snapshot_list:
                    vm_zfs_snapshot_list.remove(rsnapshot_value)
        """ In case of future debugging """
        # print("Snapshots to transfer:")
        # for item in vm_zfs_snapshot_list:
        # print(item)

        # Start the replication
        if len(remote_zfs_snapshot_list) > 0:
            print(" 🔷 DEBUG: Starting the replication operation for: '" + vm_dataset + "'")
            for snapshot_index, snapshot_value in enumerate(vm_zfs_snapshot_list):
                if snapshot_index != len(vm_zfs_snapshot_list) - 1:
                    # FIND OUT THE SIZE OF A SNAPSHOT TO TRANSFER
                    command = "zfs send -nvi " + snapshot_value + " " + vm_zfs_snapshot_list[snapshot_index + 1]
                    shell_output = subprocess.check_output(command, shell=True)
                    shell_output = shell_output.decode("UTF-8").strip("\n").split()[-1]

                    # SIZE CONVERSION TO PRINT
                    if re.match(".*G", shell_output):
                        str_shell_output = shell_output
                    elif re.match(".*M", shell_output):
                        str_shell_output = shell_output
                    elif re.match(".*K", shell_output):
                        str_shell_output = shell_output
                    else:
                        str_shell_output = shell_output + "B"

                    print(" 🔷 DEBUG: Sending INCREMENTAL snapshot " + str(snapshot_index + 1) + " out of " + str(
                        len(vm_zfs_snapshot_list) - 1) + " || (" + vm_zfs_snapshot_list[
                              snapshot_index + 1] + ", size: " + str_shell_output + ")")

                    # SIZE CONVERSION TO BYTES
                    if re.match(".*G", shell_output):
                        shell_output = float(shell_output.strip("G")) * 1024 * 1024 * 1024
                    elif re.match(".*M", shell_output):
                        shell_output = float(shell_output.strip("M")) * 1024 * 1024
                    elif re.match(".*K", shell_output):
                        shell_output = float(shell_output.strip("K")) * 1024
                    else:
                        shell_output = float(shell_output)

                    # FORCE RECEIVE IF IT'S THE FIRST SNAPSHOT IN THIS ITERATION
                    if snapshot_index == 0:
                        # print("DEBUG FISRT SNAP TO SEND!")
                        zfs_receive_command = " zfs receive -F "
                    else:
                        zfs_receive_command = " zfs receive "

                    command = "zfs send -i " + snapshot_value + " " + vm_zfs_snapshot_list[
                        snapshot_index + 1] + " | pv -p -e -r -W -t -s " + str(
                        round(shell_output)) + " | ssh " + ep_address + zfs_receive_command + vm_dataset
                    # print(command)
                    subprocess.run(command, shell=True)
                    # with subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, bufsize=2) as sp:
                    #     for line in sp.stdout:
                    #         print("Python Line! ")
                    #         print(line.decode("UTF-8").strip("\n"))
            print(" 🟢 INFO: Replication operation: done sending '" + vm_dataset + "'")
        else:
            # FIND OUT THE SIZE OF A SNAPSHOT TO TRANSFER
            command = "zfs send -nv " + vm_zfs_snapshot_list[0]
            shell_output = subprocess.check_output(command, shell=True)
            shell_output = shell_output.decode("UTF-8").strip("\n").split()[-1]

            # SIZE CONVERSION TO PRINT
            if re.match(".*G", shell_output):
                str_shell_output = shell_output
            elif re.match(".*M", shell_output):
                str_shell_output = shell_output
            elif re.match(".*K", shell_output):
                str_shell_output = shell_output
            else:
                str_shell_output = shell_output + "B"

            print(
                " 🔷 DEBUG: Starting the INITIAL replication operation for: '" + vm_dataset + "'" + " (size: " + str_shell_output + ")")

            # SIZE CONVERSION TO BYTES
            if re.match(".*G", shell_output):
                shell_output = float(shell_output.strip("G")) * 1024 * 1024 * 1024
            elif re.match(".*M", shell_output):
                shell_output = float(shell_output.strip("M")) * 1024 * 1024
            elif re.match(".*K", shell_output):
                shell_output = float(shell_output.strip("K")) * 1024
            else:
                shell_output = float(shell_output)

            command = "zfs send " + vm_zfs_snapshot_list[0] + " | pv -p -e -r -W -t -s " + str(
                round(shell_output)) + " | ssh " + ep_address + " zfs receive " + vm_dataset
            subprocess.run(command, shell=True)
            print(" 🟢 INFO: Initial snapshot replication operation: done sending '" + vm_dataset + "'")


class CloudInit:
    def __init__(self, vm_name, vm_folder, vm_ssh_keys, os_type, ip_address, network_bridge_address,
                 root_password, user_password, mac_address, new_vm_name=False, old_zfs_ds=False, new_zfs_ds=False,
                 os_comment=False):

        self.vm_name = vm_name
        self.vm_folder = vm_folder
        self.new_vm_name = new_vm_name

        self.old_zfs_ds = old_zfs_ds
        self.new_zfs_ds = new_zfs_ds

        self.output_dict = {}
        self.output_dict["random_instanse_id"] = random_password_generator(length=5)
        self.output_dict["vm_name"] = vm_name
        self.output_dict["mac_address"] = mac_address
        self.output_dict["os_type"] = os_type
        self.output_dict["ip_address"] = ip_address
        self.output_dict["network_bridge_address"] = network_bridge_address
        self.output_dict["vm_ssh_keys"] = vm_ssh_keys
        self.output_dict["root_password"] = root_password
        self.output_dict["user_password"] = user_password
        self.output_dict["os_comment"] = os_comment

    def rename(self):
        # Check if VM exists
        # vm_name = self.vm_name
        new_vm_name = self.new_vm_name

        old_zfs_ds = self.old_zfs_ds
        new_zfs_ds = self.new_zfs_ds

        new_vm_folder = self.vm_folder
        output_dict = self.output_dict
        output_dict["vm_name"] = new_vm_name

        cloud_init_files_folder = new_vm_folder + "/cloud-init-files"
        if not os.path.exists(cloud_init_files_folder):
            sys.exit(" ⛔ CRITICAL: CloudInit folder doesn't exist here: /" + old_zfs_ds)

        # Read Cloud Init Metadata
        with open("./templates/cloudinit/meta-data", "r") as file:
            md_template = file.read()
        # Render Cloud Init Metadata Template
        md_template = Template(md_template)
        md_template = md_template.render(output_dict=output_dict)
        # Write Cloud Init Metadata Template
        with open(cloud_init_files_folder + "/meta-data", "w") as file:
            file.write(md_template)

        # Create ISO file
        command = "genisoimage -output " + new_vm_folder + "/seed.iso -volid cidata -joliet -rock " + cloud_init_files_folder + "/user-data " + cloud_init_files_folder + "/meta-data " + cloud_init_files_folder + "/network-config"
        # print(command)
        subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)

        # Rename ZFS dataset
        # Check if VM is Live
        command = "zfs rename " + old_zfs_ds + " " + new_zfs_ds
        # print(command)
        subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)

    @staticmethod
    def reset(vm_name, ip_address, vm_folder, network_bridge_address, vm_ssh_keys, root_password, user_password):
        output_dict = {}
        output_dict["vm_name"] = vm_name
        output_dict["random_instanse_id"] = random_password_generator(length=5)
        output_dict["ip_address"] = ip_address
        output_dict["network_bridge_address"] = network_bridge_address
        output_dict["vm_ssh_keys"] = vm_ssh_keys
        output_dict["root_password"] = root_password
        output_dict["user_password"] = user_password

        cloud_init_files_folder = vm_folder + "/cloud-init-files"
        if not os.path.exists(cloud_init_files_folder):
            sys.exit(" ⛔ CRITICAL: CloudInit folder doesn't exist here: " + vm_folder)

        with open("./templates/vm_config_template.json", "r") as file:
            template = file.read()

        # Render VM template
        template = Template(template)
        template = template.render(output_dict)
        # Write VM template
        with open(vm_folder + "vm_config.json", "w") as file:
            file.write(template)

        # Read Cloud Init Metadata
        with open("./templates/cloudinit/meta-data", "r") as file:
            md_template = file.read()
        # Render Cloud Init Metadata Template
        md_template = Template(md_template)
        md_template = md_template.render(output_dict)
        # Write Cloud Init Metadata Template
        with open(cloud_init_files_folder + "/meta-data", "w") as file:
            file.write(md_template)

        # Read Cloud Init Network Template
        with open("./templates/cloudinit/network-config", "r") as file:
            nw_template = file.read()
        # Render Cloud Init Network Template
        nw_template = Template(nw_template)
        nw_template = nw_template.render(output_dict)
        # Write Cloud Init Network
        with open(cloud_init_files_folder + "/network-config", "w") as file:
            file.write(nw_template)

        # Read Cloud Init User Template
        with open("./templates/cloudinit/user-data", "r") as file:
            usr_template = file.read()
        # Render loud Init User Template
        usr_template = Template(usr_template)
        usr_template = usr_template.render(output_dict)
        # Write Cloud Init User Template
        with open(cloud_init_files_folder + "/user-data", "w") as file:
            file.write(usr_template)

        # Create ISO file
        command = "genisoimage -output " + vm_folder + "/seed.iso -volid cidata -joliet -rock " + cloud_init_files_folder + "/user-data " + cloud_init_files_folder + "/meta-data " + cloud_init_files_folder + "/network-config"
        subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)

    def deploy(self):
        new_vm_folder = self.vm_folder
        output_dict = self.output_dict

        cloud_init_files_folder = new_vm_folder + "/cloud-init-files"
        if not os.path.exists(cloud_init_files_folder):
            os.mkdir(cloud_init_files_folder)

        # Read Cloud Init Metadata
        with open("./templates/cloudinit/meta-data", "r") as file:
            md_template = file.read()
        # Render Cloud Init Metadata Template
        md_template = Template(md_template)
        md_template = md_template.render(output_dict=output_dict)
        # Write Cloud Init Metadata Template
        with open(cloud_init_files_folder + "/meta-data", "w") as file:
            file.write(md_template)

        # Read Cloud Init Network Template
        with open("./templates/cloudinit/network-config", "r") as file:
            nw_template = file.read()
        # Render Cloud Init Network Template
        nw_template = Template(nw_template)
        nw_template = nw_template.render(output_dict=output_dict)
        # Write Cloud Init Network
        with open(cloud_init_files_folder + "/network-config", "w") as file:
            file.write(nw_template)

        # Read Cloud Init User Template
        with open("./templates/cloudinit/user-data", "r") as file:
            usr_template = file.read()
        # Render loud Init User Template
        usr_template = Template(usr_template)
        usr_template = usr_template.render(output_dict=output_dict)
        # Write Cloud Init User Template
        with open(cloud_init_files_folder + "/user-data", "w") as file:
            file.write(usr_template)

        # Create ISO file
        command = "genisoimage -output " + new_vm_folder + "/seed.iso -volid cidata -joliet -rock " + cloud_init_files_folder + "/user-data " + cloud_init_files_folder + "/meta-data " + cloud_init_files_folder + "/network-config"
        subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)


""" Section below is responsible for the CLI input/output """
app = typer.Typer(context_settings=dict(max_content_width=800))


# app.add_typer(vmdeploy.app, name="deploy", help="Manage users in the app.")
# app.add_typer(vmlist.app, name="list")


@app.command()
def list(json: bool = typer.Option(False, help="Output json instead of a table")):
    """ List the VMs using table or JSON output """
    if json:
        print(VmList().json_output())
    else:
        VmList().table_output()


@app.command()
def info(vm_name: str = typer.Argument(..., help="Print VM config file to the screen")):
    """ Show VM info in the form of JSON output """
    vm_info_dict = VmConfigs(vm_name).vm_config_read()
    vm_info_json = json.dumps(vm_info_dict, indent=2)
    print(vm_info_json)


@app.command()
def edit(vm_name: str = typer.Argument(..., help="Edit VM config file with nano")):
    """ Manually edit the VM config file (with 'nano') """
    VmConfigs(vm_name).vm_config_manual_edit()


@app.command()
def diskexpand(vm_name: str = typer.Argument(..., help="VM name"),
               size: int = typer.Option(10, help="Number or Gigabytes to add"),
               disk: str = typer.Option("disk0.img", help="Disk image file name"),
               ):
    """ Expand VM drive. Example: hoster vm diskexpand test-vm-1 --disk disk1.img --size 100 """

    if vm_name in VmList().plainList:
        # DEBUG
        # print("All good. VM exists.")
        if CoreChecks(vm_name=vm_name, disk_image_name=disk).disk_exists():
            # DEBUG
            # print("All good. Disk exists.")
            disk_location = CoreChecks(vm_name=vm_name, disk_image_name=disk).disk_location()
            shell_command = "truncate -s +" + str(size) + "G " + disk_location
            subprocess.run(shell_command, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            print(" 🟢 INFO: Disk " + disk + "was enlarged by " + str(
                size) + "G. Reboot or start the VM to apply new settings: " + vm_name)
        else:
            sys.exit(" 🚦 ERROR: Sorry, could not find the disk: " + disk)
    else:
        sys.exit(" 🚦 ERROR: Sorry, could not find the VM with such name: " + vm_name)


@app.command()
def rename(vm_name: str = typer.Argument(..., help="VM Name"),
           new_name: str = typer.Option(..., help="New VM Name"),
           ):
    """ Rename the VM """

    if vm_name not in VmList().plainList:
        sys.exit(" 🚦 ERROR: This VM doesn't exist: " + vm_name)
    elif CoreChecks(vm_name=vm_name).vm_is_live():
        sys.exit(" 🚦 ERROR: VM is live! Please turn it off first: hoster vm stop " + vm_name)

    vm_folder = CoreChecks(vm_name=vm_name).vm_folder()
    vm_dataset = CoreChecks(vm_name=vm_name).vm_dataset()
    old_zfs_ds = vm_dataset + "/" + vm_name
    new_zfs_ds = vm_dataset + "/" + new_name

    vm_ssh_keys = []
    os_type = ""
    ip_address = ""
    network_bridge_address = ""
    root_password = ""
    user_password = ""
    mac_address = ""

    cloud_init = CloudInit(vm_name=vm_name, vm_folder=vm_folder, vm_ssh_keys=vm_ssh_keys, os_type=os_type,
                           ip_address=ip_address,
                           network_bridge_address=network_bridge_address, root_password=root_password,
                           user_password=user_password, mac_address=mac_address,
                           new_vm_name=new_name, old_zfs_ds=old_zfs_ds, new_zfs_ds=new_zfs_ds)

    cloud_init.rename()

    # Reload DNS
    VmDeploy().dns_registry()

    # Let user know, that everything went well
    print(" 🟢 INFO: VM was renamed successfully, from " + vm_name + " to " + new_name)


@app.command()
def console(vm_name: str = typer.Argument(..., help="VM Name")):
    """ Connect to VM's console """

    if vm_name not in VmList().plainList:
        sys.exit("VM doesn't exist on this system.")
    elif CoreChecks(vm_name).vm_is_live():
        command = "tmux ls | grep -c " + vm_name + " || true"
        shell_command = subprocess.check_output(command, shell=True)
        tmux_sessions = shell_command.decode("utf-8").split()[0]
        if tmux_sessions != "no server running on /tmp/tmux-0/default":
            if int(tmux_sessions) > 0:
                command = 'tmux a -t ' + '"' + vm_name + '"'
                shell_command = subprocess.check_output(command, shell=True)
            else:
                command = 'tmux new-session -s ' + vm_name + ' "cu -l /dev/nmdm-' + vm_name + '-1B"'
                subprocess.run(command, shell=True)
    else:
        sys.exit("VM is not running. Start the VM first to connect to it's console.")


@app.command()
def destroy(vm_name: str = typer.Argument(..., help="VM Name"),
            force: bool = typer.Option(False, help="Kill and destroy the VM, even if it's running"),
            ):
    """ Completely remove the VM from this system! """
    Operation.destroy(vm_name=vm_name)

    # Reload DNS
    VmDeploy().dns_registry()


@app.command()
def destroy_all(force: bool = typer.Option(False, help="Kill and destroy all VMs, even if they are running")):
    """ Completely remove all VMs from this system! """
    vm_list = VmList().plainList
    for _vm in vm_list:
        Operation.destroy(vm_name=_vm, force=force)

    # Let user know that he can remove deployment snapshots
    print()
    print(" 🔶 INFO: Execute this command to find and manually remove old deployment snapshots:")
    print("          zfs list -t all | grep \"@deployment_\" | awk '{ print $1 }'")
    print(" 🔶 INFO: Or execute this command to find and automatically remove old test deployment snapshots:")
    print(
        "          for ITEM in $(zfs list -t all | grep \"@deployment_\" | awk '{ print $1 }' | grep test-vm-); do zfs destroy $ITEM; done")
    print()

    # Reload DNS
    VmDeploy().dns_registry()


@app.command()
def snapshot(vm_name: str = typer.Argument(..., help="VM Name"),
             stype: str = typer.Option("custom", help="Snapshot type: daily, weekly, etc"),
             keep: int = typer.Option(3, help="How many snapshots to keep")
             ):
    """
    Snapshot the VM (RAM snapshots are not supported). Snapshot will be taken at the storage level: ZFS or GlusterFS.
    Example: hoster vm snapshot test-vm-1 --type weekly --keep 5
    """

    Operation.snapshot(vm_name=vm_name, stype=stype, keep=keep)


@app.command()
def snapshot_all(stype: str = typer.Option("custom", help="Snapshot type: daily, weekly, etc"),
                 keep: int = typer.Option(3, help="How many snapshots to keep")
                 ):
    """ Snapshot all VMs """

    vm_list = VmList().plainList
    for _vm in vm_list:
        vm_prod_status_local = CoreChecks(vm_name=_vm).vm_in_production()
        vm_live_status_local = CoreChecks(vm_name=_vm).vm_is_live()
        if vm_prod_status_local and vm_live_status_local:
            Operation.snapshot(vm_name=_vm, keep=keep, stype=stype)


@app.command()
def kill(vm_name: str = typer.Argument(..., help="VM Name")):
    """ Kill the VM immediately! """
    Operation.kill(vm_name=vm_name)


@app.command()
def kill_all():
    """ Kill all VMs on this system! """

    vm_list = VmList().plainList
    for _vm in vm_list:
        Operation.kill(vm_name=_vm)


@app.command()
def start(vm_name: str = typer.Argument(..., help="VM name"),
          ):
    """ Power on the VM """

    Operation.start(vm_name=vm_name)


@app.command()
def start_all(wait: int = typer.Option(5, help="Seconds to wait before starting the next VM on the list")
              ):
    """ Power on all production VMs """

    vm_list = VmList().plainList
    for _vm in vm_list:
        if not CoreChecks(vm_name=_vm).vm_is_live():
            if CoreChecks(vm_name=_vm).vm_in_production():
                Operation.start(vm_name=_vm)
                time.sleep(wait)
                wait = wait + 3
        else:
            print("VM is already live: " + _vm)


@app.command()
def stop(vm_name: str = typer.Argument(..., help="VM name"),
         ):
    """ Gracefully stop the VM """
    Operation.stop(vm_name=vm_name)


@app.command()
def stop_all(wait: int = typer.Option(5, help="Seconds to wait before stopping the next VM on the list")
             ):
    """ Gracefully stop all VMs running on this system """

    vm_list = VmList().plainList
    for vm in vm_list:
        if CoreChecks(vm_name=vm).vm_is_live():
            Operation.stop(vm_name=vm)
            time.sleep(wait)
        else:
            print("VM is already stopped: " + vm)


@app.command()
def show_log(vm_name: str = typer.Argument(..., help="VM name"),
             ):
    """ Show the live VM's log output """
    Operation.show_log(vm_name=vm_name)


@app.command()
def deploy(vm_name: str = typer.Argument("test-vm", help="New VM name"),
           os_type: str = typer.Option("ubuntu2004", help="OS Type, for example: debian11 or ubuntu2004"),
           # ip_address:str = typer.Option("10.0.0.0", help="Specify the IP address or leave at default to generate a random address"),
           ds_id: int = typer.Option(0, help="Dataset ID to which this VM will be deployed"),
           ):
    """ New VM deployment """

    deployment_output = VmDeploy(vm_name=vm_name, os_type=os_type, dataset_id=ds_id).deploy()
    # Reload DNS
    VmDeploy().dns_registry()
    # Let user know, that everything went well
    print(" 🟢 INFO: VM was deployed successfully: " + deployment_output["vm_name"])


@app.command()
def cireset(vm_name: str = typer.Argument(..., help="VM name"),
            ):
    """ Reset the VM settings, including passwords, network settings, user keys, etc. """
    if vm_name not in VmList().plainList:
        sys.exit(" 🚦 ERROR: This VM doesn't exist: " + vm_name)
    elif CoreChecks(vm_name=vm_name).vm_is_live():
        sys.exit(" 🚦 ERROR: VM is live! Please turn it off first: hoster vm stop " + vm_name)
    vm_config_dict = VmConfigs(vm_name).vm_config_read()
    # print(vm_config_dict)
    vm_folder = CoreChecks(vm_name=vm_name).vm_folder()
    # print(vm_folder)
    # _ Load host config _#
    with open("./configs/host.json", "r") as file:
        host_file = file.read()

    host_dict = json.loads(host_file)
    # print(host_dict)

    host_name = host.get_hostname()
    # print(host_name)

    # _ Load networks config _#
    with open("./configs/networks.json", "r") as file:
        networks_file = file.read()
    networks_dict = json.loads(networks_file)
    network_bridge_name = networks_dict["networks"][0]["bridge_name"]
    # print(network_bridge_name)

    network_ip_address = ip_address_generator()
    # print(network_ip_address)

    vm_ssh_keys = []
    host_ssh_keys = []
    if vm_config_dict["include_hostwide_ssh_keys"]:
        key_index = 0
        for _key in host_dict["host_ssh_keys"]:
            _ssh_key = {}
            _ssh_key["key_value"] = _key["key_value"]
            _ssh_key["key_owner"] = host_name
            _ssh_key["comment"] = "Host SSH key"
            key_index = key_index + 1
            host_ssh_keys.append(_ssh_key)
    for _key in vm_config_dict["vm_ssh_keys"]:
        _ssh_key = {}
        _ssh_key["key_value"] = _key["key_value"]
        _ssh_key["key_owner"] = _key["key_owner"]
        _ssh_key["comment"] = _key["comment"]
        vm_ssh_keys.append(_ssh_key)
    # print(vm_ssh_keys)

    for _host_ssh_key in host_ssh_keys:
        if _host_ssh_key not in vm_ssh_keys:
            vm_ssh_keys.append(_host_ssh_key)

    vnc_port = VmDeploy.vm_vnc_port_generator()
    # print(vnc_port)

    vm_config_dict["parent_host"] = host_name
    vm_config_dict["networks"][0]["ip_address"] = network_ip_address
    vm_config_dict["networks"][0]["network_bridge"] = network_bridge_name
    vm_config_dict["vm_ssh_keys"] = vm_ssh_keys
    vm_config_dict["vnc_port"] = str(vnc_port)

    final_output = json.dumps(vm_config_dict, indent=3)

    # Write VM template
    vm_folder = CoreChecks(vm_name=vm_name).vm_folder()
    # print(vm_folder)
    with open(vm_folder + "/vm_config.json", "w") as file:
        file.write(final_output)

    # cloud_init.reset()
    cloud_init_files_folder = vm_folder + "/cloud-init-files"
    if not os.path.exists(cloud_init_files_folder):
        sys.exit(" ⛔ CRITICAL: CloudInit folder doesn't exist at this location: " + vm_folder)

    output_dict = {"random_instanse_id": random_password_generator(length=5), "vm_name": vm_name,
                   "mac_address": vm_config_dict["networks"][0]["network_mac"], "os_type": vm_config_dict["os_type"],
                   "ip_address": vm_config_dict["networks"][0]["ip_address"],
                   "network_bridge_address": networks_dict["networks"][0]["bridge_address"]}

    ci_vm_ssh_keys = []
    for _ssh_key in vm_ssh_keys:
        ci_vm_ssh_keys.append(_ssh_key["key_value"])
    output_dict["vm_ssh_keys"] = ci_vm_ssh_keys

    output_dict["root_password"] = random_password_generator(capitals=True, numbers=True, length=53)
    output_dict["user_password"] = random_password_generator(capitals=True, numbers=True, length=53)

    # Read Cloud Init Metadata
    with open("./templates/cloudinit/meta-data", "r") as file:
        md_template = file.read()
    # Render Cloud Init Metadata Template
    md_template = Template(md_template)
    md_template = md_template.render(output_dict=output_dict)
    # Write Cloud Init Metadata Template
    with open(cloud_init_files_folder + "/meta-data", "w") as file:
        file.write(md_template)

    # Read Cloud Init Network Template
    with open("./templates/cloudinit/network-config", "r") as file:
        nw_template = file.read()
    # Render Cloud Init Network Template
    nw_template = Template(nw_template)
    nw_template = nw_template.render(output_dict=output_dict)
    # Write Cloud Init Network
    with open(cloud_init_files_folder + "/network-config", "w") as file:
        file.write(nw_template)

    # Read Cloud Init User Template
    with open("./templates/cloudinit/user-data", "r") as file:
        usr_template = file.read()
    # Render loud Init User Template
    usr_template = Template(usr_template)
    usr_template = usr_template.render(output_dict=output_dict)
    # Write Cloud Init User Template
    with open(cloud_init_files_folder + "/user-data", "w") as file:
        file.write(usr_template)

    # Create ISO file
    command = "genisoimage -output " + vm_folder + "/seed.iso -volid cidata -joliet -rock " + cloud_init_files_folder + "/user-data " + cloud_init_files_folder + "/meta-data " + cloud_init_files_folder + "/network-config"
    subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)

    # Reload DNS
    VmDeploy().dns_registry()

    # Let user know, that everything went well
    print(" 🟢 INFO: VM was reset successfully: " + vm_name)


@app.command()
def replicate(vm_name: str = typer.Argument(..., help="VM name"),
              ep_address: str = typer.Option("192.168.120.18", help="Endpoint server address, i.e. 192.168.1.1"),
              ep_port: str = typer.Option("22", help="Endpoint server SSH port"),
              direction: str = typer.Option("push", help="Direction of the replication: push or pull")
              ):
    """ Replicate the VM to another host """

    if direction == "push":
        ZFSReplication.push(vm_name=vm_name, ep_address=ep_address, ep_port=ep_port)
    elif direction == "pull":
        print("This function has not been implemented yet!")
    else:
        print("Only available options are \"pull\" and \"push\"!")


@app.command()
def replicate_all(ep_address: str = typer.Option("192.168.120.18", help="Endpoint server address, i.e. 192.168.1.1"),
                  ep_port: str = typer.Option("22", help="Endpoint server SSH port"),
                  direction: str = typer.Option("push", help="Direction of the replication: push or pull")
                  ):
    """ Replicate all production VMs to another host """

    vm_list = VmList().plainList
    for vm in vm_list:
        vm_live_status = CoreChecks(vm_name=vm).vm_cpus()["live_status"]
        if vm_live_status == "production":
            if direction == "push":
                ZFSReplication.push(vm_name=vm, ep_address=ep_address, ep_port=ep_port)
            elif direction == "pull":
                print("This function has not been implemented yet!")
            else:
                print("Only available options are \"pull\" and \"push\"!")


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
