import os
import re
import sys
from time import sleep

from rich.console import Console
# from rich.panel import Panel
import typer
import invoke
import json

# Own libraries
from cli.vm import vm
from cli.host import host
from cli.network import network
from cli.dataset import dataset

""" Section below is responsible for the CLI input/output """
app = typer.Typer(context_settings=dict(max_content_width=800))
app.add_typer(vm.app, name="vm", help="List, Create, Remove or any other VM related operations")
app.add_typer(host.app, name="host", help="Show or modify host related information")
app.add_typer(network.app, name="network", help="Show or modify network related information")
app.add_typer(dataset.app, name="dataset", help="Show or modify storage dataset related information")


@app.command()
def version(json_plain: bool = typer.Option(False, help="Plain JSON output")):
    """ Show version and exit """

    version_string = "development-0.1-alpha"
    if json_plain:
        dict_output = {"version": version_string}
        json_output = json.dumps(dict_output, sort_keys=False)
        print(json_output)
    else:
        print("Version: " + version_string)


@app.command()
def init():
    """ Initialise all modules and services required by 'hoster' """

    host.init()
    network.init()


@app.command()
def self_update():
    """ Pull the latest updates from our Git repo """

    hoster_red_folder = "/opt/hoster-red/"
    if os.path.exists(hoster_red_folder):
        os.chdir(hoster_red_folder)
    else:
        Console(stderr=True).print("Hoster folder doesn't exist!")
        sys.exit(1)

    with Console().status("[bold royal_blue1]Pulling the latest changes...[/]"):
        invoke.run("git reset --hard", hide=True)
        try:
            git_result = invoke.run("git pull", hide=True)
            git_output = git_result.stdout.splitlines()
            re_out_1 = re.compile(".*Already up to date.*")
            re_out_2 = re.compile(".*Already up-to-date.*")
            for index, value in enumerate(git_output):
                if re_out_1.match(value) or re_out_2.match(value):
                    git_pull_job_status = " 🟢 INFO: [green]Hoster is already up-to-date![/]"
                elif not (re_out_1.match(value) or re_out_2.match(value)) and (index + 1) == len(git_output):
                    git_pull_job_status = " 🟢 INFO: [green]Hoster was updated successfully![/]"
        except invoke.exceptions.UnexpectedExit as e:
            pass
    Console().print(git_pull_job_status)

    with Console().status("[bold royal_blue1]Upgrading PIP dependencies...[/]"):
        command = hoster_red_folder + "venv/bin/python3 -m pip install -r requirements.txt --upgrade"
        invoke.run(command, hide=True)
        pip_upgrade_job_status = " 🟢 INFO: [green]Hoster PIP dependencies were updated successfully![/]"
    Console().print(git_pull_job_status)


@app.callback(invoke_without_command=True)
def main(ctx: typer.Context):
    """ Bhyve automation framework """

    if ctx.invoked_subcommand is None:
        print()
        host.table_output(table_title=True)
        print()
        network.info(table=True, table_title=True, json_pretty=False, json_plain=False)
        print()
        dataset.info(table_title=True, json_pretty=False, json_plain=False)
        print()
        vm.VmList().table_output(table_title=True)
        print()


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
