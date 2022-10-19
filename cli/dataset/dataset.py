import json
import os

from rich.console import Console
from rich.table import Table
from rich import box
import typer


class DatasetList:
    def __init__(self):
        with open("./configs/datasets.json", "r") as file:
            datasets_file = file.read()
        datasets_file = json.loads(datasets_file)
        self.datasets = datasets_file

    def json_output(self):
        datasets_json = json.dumps(self.datasets, indent=2)
        return datasets_json


""" Section below is responsible for the CLI input/output """
app = typer.Typer()


@app.command()
def info(
        json: bool = typer.Option(False, help="Output json instead of a table"),
        table_title: bool = typer.Option(False, help="Show table title (useful when showing multiple tables)")
):
    if not table_title:
        table = Table(box=box.ROUNDED)
    else:
        table = Table(title=" Host Information", box=box.ROUNDED, title_justify="left")
    ds = DatasetList()
    print(ds)


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
