from rich.console import Console
from cli.host import host
import subprocess
import requests
import typer
import json
import sys
import os


class NebulaFuncs:
    def __init__(self):
        os.chdir("/opt/hoster-red/")
        if not os.path.exists("./configs/nebula.json"):
            print(" 🟢 FATAL: Nebula config was not found!")
            sys.exit(1)

        with open("./configs/nebula.json", "r") as f:
            nebula_config_file = f.read()
        nebula_config_dict = json.loads(nebula_config_file)

        # CLUSTER AUTH
        self.host_name = host.get_hostname()
        self.cluster_name = nebula_config_dict.get("cluster_name")
        self.cluster_id = nebula_config_dict.get("cluster_id")
        self.host_id = nebula_config_dict.get("host_id")

        # LOCAL SETTINGS
        self.nat_punch = nebula_config_dict.get("nat_punch")
        self.listen_address = nebula_config_dict.get("listen_address")
        self.listen_port = nebula_config_dict.get("listen_port")
        self.mtu = nebula_config_dict.get("mtu")
        self.use_relays = nebula_config_dict.get("use_relays")
        self.api_server = nebula_config_dict.get("api_server")

        if not os.path.exists("/opt/nebula/"):
            os.mkdir("/opt/nebula/")
            subprocess.run("chmod 700 /opt/nebula/", shell=True)

        self.config_text_local = "/opt/nebula/config.yml"

    def get_certs(self, reload: bool = True):
        certificate_request_url = "https://" + self.api_server + "/get_certs?cluster_name=" + self.cluster_name + \
                                  "&cluster_id=" + self.cluster_id + "&host_name=" + self.host_name + "&host_id=" + self.host_id
        certificate_request = requests.get(certificate_request_url)
        if certificate_request.status_code == 200:
            with open("cert_files.sh", "w") as f:
                f.write(certificate_request.text)
            command = "bash cert_files.sh"
            subprocess.run(command, shell=True)
        else:
            Console().print(" 🚦 FATAL: API server is refusing your request! Check your nebula.json config!")
            return

    def get_latest_service_file(self, reload: bool = True):
        service_request_url = "https://" + self.api_server + "/get_bins?os=freebsd&nebula=false&service=true"
        service_request = requests.get(service_request_url)
        if service_request.status_code == 200:
            with open("/opt/nebula/nebula_service.sh", "w") as f:
                f.write(service_request.text)
            command = "chmod +x /opt/nebula/nebula_service.sh"
            subprocess.run(command, shell=True)
        else:
            Console().print(" 🚦 FATAL: API server is refusing your request! Check your nebula.json config!")
            return

        if reload:
            command = "/opt/nebula/nebula_service.sh"
            subprocess.run(command, shell=True)
            Console().print(" 🟢 INFO: New service file has been installed and reload initiated")
        else:
            Console().print(" 🟢 INFO: New service file has been downloaded")

    def get_latest_nebula_bin(self, reload: bool = True):
        with Console().status("[royal_blue1]Downloading the latest Nebula binary...[/]"):
            binary_request_url = "https://" + self.api_server + "/get_bins?os=freebsd&nebula=true&service=false"
            binary_request = requests.get(binary_request_url, stream=True)
            if binary_request.status_code == 200:
                subprocess.run("kill \"$(pgrep -lf '/opt/nebula/config.yml' | awk '{ print $1 }')\"", shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)
                if os.path.exists("/opt/nebula/nebula"):
                    os.remove("/opt/nebula/nebula")
                with open("/opt/nebula/nebula", "wb") as f:
                    for data in binary_request.iter_content(chunk_size=8):
                        f.write(data)
            else:
                Console().print(" 🚦 FATAL: API server is refusing your request! Check your nebula.json config!")
                return

        command = "chmod +x /opt/nebula/nebula"
        subprocess.run(command, shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)
        if reload:
            command = "/opt/nebula/nebula_service.sh"
            subprocess.run(command, shell=True)
            Console().print(" 🟢 INFO: New binary has been installed and service reloaded")
        else:
            Console().print(" 🟢 INFO: New binary has been installed")

    def get_config(self, reload: bool = True):
        config_request_url = "https://" + self.api_server + "/get_config?cluster_name=" + self.cluster_name + "&cluster_id="\
                             + self.cluster_id + "&host_name=" + self.host_name + "&host_id=" + self.host_id + "&nat_punch=" \
                             + self.nat_punch + "&listen_host=" + self.listen_address + "&listen_port=" + self.listen_port \
                             + "&mtu=" + self.mtu + "&use_relays=" + self.use_relays

        config_request = requests.get(config_request_url)
        if config_request.status_code == 200:
            config_text = config_request.text

            if os.path.exists(self.config_text_local):
                with open(self.config_text_local, "r") as f:
                    config_text_local_var = f.read()
            else:
                config_text_local_var = ""

            if config_text != config_text_local_var:
                print(" 🔷 DEBUG: Config file was changed! Downloading new config")
                with open(self.config_text_local, "w") as f:
                    f.write(config_text)
                print(" 🔷 DEBUG: Downloading new certificates")
                self.get_certs()
            else:
                print(" 🔷 DEBUG: Config file has not changed, skipping the download step...")
            if reload:
                print(" 🔷 DEBUG: Reloading/starting the service")
                command = "/opt/nebula/nebula_service.sh noout"
                subprocess.run(command, shell=True)
                print(" 🟢 INFO: All done. Welcome to the cluster, buddy!")
        else:
            Console().print(" 🚦 FATAL: API server is refusing your request! Check your nebula.json config!")
            return


""" Section below is responsible for the CLI input/output """
app = typer.Typer()


@app.command()
def init(service_reload: bool = typer.Option(True, help="Reload the service after initialisation"),
         download_binary: bool = typer.Option(True, help="Download a fresh version of Nebula binary"),
         download_service: bool = typer.Option(True, help="Download a fresh version of Nebula service file"),
         ):
    """ Initialize Nebula on this hoster (download, setup and configure) """
    if download_service:
        NebulaFuncs().get_latest_service_file(reload=False)
    if download_binary:
        NebulaFuncs().get_latest_nebula_bin(reload=False)
    NebulaFuncs().get_config(reload=service_reload)


@app.command()
def update_binary(service_reload: bool = typer.Option(True, help="Reload the service after initialisation"),
                  ):
    """ Download the latest compatible Nebula binary """
    NebulaFuncs().get_latest_nebula_bin(reload=service_reload)


@app.command()
def update_service(service_reload: bool = typer.Option(True, help="Reload the service after initialisation"),
                   ):
    """ Download the latest Nebula service file """
    NebulaFuncs().get_latest_service_file(reload=service_reload)


@app.command()
def reload_service():
    """ Start or restart the Nebula service (status detected automatically, no flags needed) """
    NebulaFuncs().get_config(reload=True)


@app.command()
def kill_service():
    """ Kill the Nebula service process """
    subprocess.run("kill \"$(pgrep -lf '/opt/nebula/config.yml' | awk '{ print $1 }')\"", shell=True, stderr=subprocess.DEVNULL, stdout=subprocess.DEVNULL)
    print(" 🔶 INFO: Nebula service was killed")


@app.command()
def show_log():
    """ Show the latest Nebula logs (runs tail -f /opt/nebula/log.txt)"""
    subprocess.run("tail -f /opt/nebula/log.txt", shell=True)


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
