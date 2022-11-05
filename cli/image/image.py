import os
import sys
import json
import zipfile

import typer
import invoke
import requests
from tqdm import tqdm
from natsort import natsorted


""" Section below is responsible for the CLI input/output """
app = typer.Typer()


@app.command()
def download(
        os_type: str = typer.Argument("debian11", help="OS or distro to download"),
        chunk_size: int = typer.Option(16, help="Download file chunk size"),
        zfs_path: str = typer.Option("zroot/vm_encrypted", help="Set the ZFS dataset path"),
):
    """ Download a ready to deploy OS image """

    if zfs_path[0] == "/" or zfs_path[-1] == "/":
        print("Sorry, ZFS path can't start or end with '/'", file=sys.stderr)
        sys.exit(1)

    json_image_url = "https://images.yari.pw/"
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
    latest_os_image = natsorted(os_image_list)[0]

    image_url = "https://images.yari.pw/images/" + latest_os_image
    image_zip_name = os_type + ".zip"

    image_download_stream = requests.get(image_url, stream=True)
    image_size = int(image_download_stream.headers.get("content-length"))

    image_end_location = "/" + zfs_path + "/template-" + os_type + "/"
    if os.path.exists(image_end_location + "/disk0.img"):
        print("Sorry, the image file already exists: " + "/" + zfs_path + "/template-" + os_type + "/disk0.img", file=sys.stderr)
        sys.exit(1)
    elif os.path.exists(image_end_location):
        pass
    else:
        command = "zfs create " + zfs_path + "/template-" + os_type
        print("Executing: " + command)
        result = invoke.run(command, hide=True)

    with open("/tmp/" + image_zip_name, "wb") as handle:
        try:
            for data in tqdm(image_download_stream.iter_content(chunk_size=chunk_size), desc="Downloading " + os_type + "... ", colour="green", total=image_size, initial=0, unit="b", unit_divisor=1024, unit_scale=True):
                handle.write(data)
        except KeyboardInterrupt as e:
            print("Process was cancelled by the user (Ctrl+C)")
            os.remove("/tmp/" + image_zip_name)
            sys.exit(1)

    if zipfile.is_zipfile("/tmp/" + image_zip_name):
        with zipfile.ZipFile("/tmp/" + image_zip_name, "r") as zip_ref:
            zip_ref.extractall(image_end_location)
            os.remove("/tmp/" + image_zip_name)

    else:
        print("Sorry, the downloaded file is not a ZIP archive: " + "/tmp/" + image_zip_name, file=sys.stderr)
        sys.exit(1)


@app.command()
def update():
    """ Initialise Kernel modules and required services """
    print("This function will soon be ready...")


@app.command()
def list_all_images():
    """ Initialise Kernel modules and required services """
    print("This function will soon be ready...")


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
