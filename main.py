import typer

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
def version(json: bool = typer.Option(False, help="JSON output")):
    """ Show version and exit """
    print("Version: development-0.1-alpha")


@app.command()
def init():
    """ Initialise all modules and services required by 'hoster' """
    host.init()
    network.init()


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
