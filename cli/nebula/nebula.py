from rich.console import Console
from cli.host import host
import subprocess
import requests
import typer
import json
import sys
import os


if not os.path.exists("./configs/nebula.json"):
    print(" 🟢 FATAL: Nebula config was not found!")
    sys.exit(1)

with open("./configs/nebula.json", "r") as f:
    nebula_config_file = f.read()
nebula_config_dict = json.loads(nebula_config_file)

# CLUSTER AUTH
host_name = host.get_hostname()
cluster_name = nebula_config_dict.get("cluster_name")
cluster_id = nebula_config_dict.get("cluster_id")
host_id = nebula_config_dict.get("host_id")

# LOCAL SETTINGS
nat_punch = nebula_config_dict.get("nat_punch")
listen_address = nebula_config_dict.get("listen_address")
listen_port = nebula_config_dict.get("listen_port")
mtu = nebula_config_dict.get("mtu")
use_relays = nebula_config_dict.get("use_relays")
api_server = nebula_config_dict.get("api_server")

if not os.path.exists("/opt/nebula/"):
    os.mkdir("/opt/nebula/")
    subprocess.run("chmod 700 /opt/nebula/", shell=True)

config_text_local = "/opt/nebula/config.yaml"


def get_certs(reload: bool = True):
    certificate_request_url = "https://" + api_server + "/get_certs?cluster_name=" + cluster_name + \
                              "&cluster_id=" + cluster_id + "&host_name=" + host_name + "&host_id=" + host_id
    certificate_request = requests.get(certificate_request_url)
    if certificate_request.status_code == 200:
        with open("cert_files.sh", "w") as f:
            f.write(certificate_request.text)
        command = "bash cert_files.sh"
        subprocess.run(command, shell=True)


def get_latest_service_file(reload: bool = True):
    service_request_url = "https://" + api_server + "/get_bins?os=freebsd&nebula=false&service=true"
    service_request = requests.get(service_request_url)
    if service_request.status_code == 200:
        with open("/opt/nebula/nebula_service.sh", "w") as f:
            f.write(service_request.text)
    if reload:
        command = "chmod +x /opt/nebula/nebula_service.sh"
        subprocess.run(command, shell=True)
        Console().print(" 🟢 INFO: New service file has been installed and reload initiated")
    else:
        Console().print(" 🟢 INFO: New service file has been downloaded")


def get_latest_nebula_bin(reload: bool = True):
    with Console().status("[royal_blue1]Downloading the latest Nebula binary...[/]"):
        binary_request_url = "https://" + api_server + "/get_bins?os=freebsd&nebula=true&service=false"
        binary_request = requests.get(binary_request_url, stream=True)
        if binary_request.status_code == 200:
            subprocess.run("kill \"$(pgrep -lf '/opt/nebula/config.yml' | awk '{ print $1 }')\"", stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)
            if os.path.exists("/opt/nebula/nebula"):
                os.remove("/opt/nebula/nebula")
            with open("/opt/nebula/nebula", "wb") as f:
                for data in binary_request.iter_content():
                    f.write(data)
    command = "chmod +x /opt/nebula/nebula"
    subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)
    if reload:
        command = "/opt/nebula/nebula_service.sh"
        subprocess.run(command, shell=True)
        Console().print(" 🟢 INFO: New binary has been installed and service reloaded")
    else:
        Console().print(" 🟢 INFO: New binary has been installed")


def get_config(reload: bool = True):
    config_request_url = "https://" + api_server + "/get_config?cluster_name=" + cluster_name + "&cluster_id="\
                         + cluster_id + "&host_name=" + host_name + "&host_id=" + host_id + "&nat_punch=" \
                         + nat_punch + "&listen_host=" + listen_address + "&listen_port=" + listen_port \
                         + "&mtu=" + mtu + "&use_relays=" + use_relays

    config_request = requests.get(config_request_url)
    if config_request.status_code == 200:
        config_text = config_request.text

        if os.path.exists(config_text_local):
            with open(config_text_local, "r") as f:
                config_text_local_var = f.read()
        else:
            config_text_local_var = ""

        if config_text != config_text_local_var:
            print(" 🔷 DEBUG: Config file was changed! Downloading new config")
            with open(config_text_local, "w") as f:
                f.write(config_text)
            print(" 🔷 DEBUG: Downloading new certificates")
            get_certs()
            if reload:
                print(" 🔷 DEBUG: Reloading the service")
                command = "/opt/nebula/nebula_service.sh"
                subprocess.run(command, shell=True)
            print(" 🟢 INFO: All done, and you now have the latest Nebula settings. Welcome back to the cluster, buddy!")
        else:
            print(" 🔷 DEBUG: Config file was not changed, skipping any further steps...")


""" Section below is responsible for the CLI input/output """
app = typer.Typer()


@app.command()
def init(service_reload: bool = typer.Option(True, help="Reload the service after initialisation"),
         ):
    """ Initialize Nebula on this hoster (download, setup and configure) """
    if not os.path.exists("./configs/nebula.json"):
        print(" 🟢 FATAL: Nebula config was not found!")
        sys.exit(1)
    get_latest_service_file(reload=False)
    get_latest_nebula_bin(reload=False)
    get_config(reload=service_reload)


def update_binary(service_reload: bool = typer.Option(True, help="Reload the service after initialisation"),
                  ):
    """ Download the latest compatible Nebula binary """
    if not os.path.exists("./configs/nebula.json"):
        print(" 🟢 FATAL: Nebula config was not found!")
        sys.exit(1)
    get_latest_nebula_bin(reload=service_reload)


def update_service(service_reload: bool = typer.Option(True, help="Reload the service after initialisation"),
                   ):
    """ Download the latest Nebula service file """
    if not os.path.exists("./configs/nebula.json"):
        print(" 🟢 FATAL: Nebula config was not found!")
        sys.exit(1)
    get_latest_service_file(reload=service_reload)


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
