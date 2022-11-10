import os
import sys
import json
import zipfile
import subprocess

import typer
import requests
from natsort import natsorted
from rich.console import Console


""" Section below is responsible for the CLI input/output """
app = typer.Typer()


@app.command()
def download(
        os_type: str = typer.Argument("debian11", help="OS or distro to download"),
        zfs_path: str = typer.Option("zroot/vm-encrypted", help="Set the ZFS dataset path"),
        image_source_url: str = typer.Option("https://images.yari.pw/", help="Set the URL where to get images from"),
        force_update: bool = typer.Option(False, help="Replace an image if it exists (useful to update old images)"),
):
    """ Download a ready to deploy OS image """

    if zfs_path[0] == "/" or zfs_path[-1] == "/":
        print("Sorry, ZFS path can't start or end with '/'", file=sys.stderr)
        sys.exit(1)

    json_image_url = image_source_url
    json_images = requests.get(json_image_url)
    images_dict = json.loads(json_images.text)

    existing_images = []
    for i in [*images_dict["vm_images"]]:
        existing_images.append(*i)

    if os_type not in existing_images:
        print("Sorry, image doesn't exist on this server: " + json_image_url, file=sys.stderr)
        print("List of available images: " + str(existing_images), file=sys.stderr)
        sys.exit(1)

    os_image_list = []
    for i in images_dict["vm_images"]:
        temp = i.get(os_type)
        if temp:
            os_image_list = temp
    latest_os_image = natsorted(os_image_list)[-1]

    image_url = "https://images.yari.pw/images/" + latest_os_image
    image_zip_name = os_type + ".zip"

    image_end_location = "/" + zfs_path + "/template-" + os_type + "/"
    if os.path.exists(image_end_location + "/disk0.img"):
        if force_update:
            os.remove(image_end_location + "/disk0.img")
        else:
            print("Sorry, the image file already exists: " + "/" + zfs_path + "/template-" + os_type + "/disk0.img", file=sys.stderr)
            sys.exit(1)
    elif os.path.exists(image_end_location):
        pass
    else:
        command = "zfs create " + zfs_path + "/template-" + os_type
        print("Executing: " + command)
        subprocess.run(command, shell=True, stdout=subprocess.DEVNULL)

    Console().print("Will download: " + image_url)
    try:
        command = "wget " + image_url + " -O /tmp/" + os_type + ".zip -q --show-progress"
        subprocess.run(command, shell=True)
    except KeyboardInterrupt as e:
        Console().print("Process was cancelled by the user (Ctrl+C)")
        os.remove("/tmp/" + image_zip_name)
        sys.exit(1)

    if zipfile.is_zipfile("/tmp/" + image_zip_name):
        with Console().status("[bold royal_blue1]Unzipping the image archive...[/]"):
            with zipfile.ZipFile("/tmp/" + image_zip_name, "r") as zip_ref:
                zip_ref.extractall(image_end_location)
                os.remove("/tmp/" + image_zip_name)
        Console().print("Image was unpacked to: /" + zfs_path + "/template-" + os_type + "/disk0.img")
        Console().print("Downloaded archive was cleaned up from: /tmp/" + image_zip_name)
    else:
        Console(stderr=True).print("Sorry, the downloaded file is not a ZIP archive: " + "/tmp/" + image_zip_name)
        os.remove("/tmp/" + image_zip_name)
        sys.exit(1)


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
