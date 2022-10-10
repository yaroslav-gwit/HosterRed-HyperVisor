# from tqdm.rich import trange, tqdm
from tqdm import tqdm, trange
import requests
import os

def get_almalinux8_image():
    url = "https://github.com/yaroslav-gwit/PyVM-Bhyve/releases/download/202109/almalinux8.zip"
    response = requests.get(url, stream=True)
    file_size = int(response.headers.get("content-length"))

    image_name = "image.img"
    with open(image_name, "wb") as handle:
        try:
            for data in tqdm(response.iter_content(), desc="Downloading AlmaLinux 8 Image", colour="green",
                                total=file_size, initial=0, unit="b", unit_divisor=1024, unit_scale=True):
                
                handle.write(data)
        except KeyboardInterrupt as e:
            print("Process was cancelled by the user (Ctrl+C)")
            os.remove(image_name)
