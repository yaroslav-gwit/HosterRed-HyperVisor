import typer
from cli.network import internal_classes as IC

""" Section below is responsible for the CLI input/output """
app = typer.Typer(context_settings=dict(max_content_width=800))


@app.command()
def init():
    """ Initialize Hoster networks """

    IC.NetworkInit().init()
