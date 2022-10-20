import json

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
def info(json_plain: bool = typer.Option(False, help="Output json instead of a table"),
         json_pretty: bool = typer.Option(False, help="Output json instead of a table"),
         table_title: bool = typer.Option(False, help="Show table title (useful when showing multiple tables)")
         ):
    datasets = DatasetList().datasets

    if not json_plain and not json_pretty:
        if not table_title:
            table = Table(box=box.ROUNDED, show_lines=True, )
        else:
            table = Table(title=" Dataset information", box=box.ROUNDED, title_justify="left", show_lines=True, )

        table.add_column("DS ID", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Name", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Type", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Mount path", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("ZFS dataset", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Is encrypted?", justify="center", style="bright_cyan", no_wrap=True)
        table.add_column("Comment", justify="center", style="bright_cyan", no_wrap=True)

        for ds in datasets["datasets"]:
            table.add_row(ds["id"], ds["name"], ds["type"], ds["mount_path"],
                          ds["zfs_path"], str(ds["encrypted"]), ds["comment"], )

        Console().print(table)

    elif json_plain:
        json_output = json.dumps(datasets)
        print(json_output)

    elif json_pretty:
        json_output = json.dumps(datasets)
        Console().print_json(json_output, indent=3, sort_keys=False)


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
