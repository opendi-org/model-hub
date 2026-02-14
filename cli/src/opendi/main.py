"""CLI entry point. Called when the user runs the `opendi` command."""

import typer

app = typer.Typer(
    name="opendi",
    help="OpenDI Model Hub CLI — cross-platform client for the OpenDI hub.",
)


@app.callback(invoke_without_command=True)
def main(ctx: typer.Context) -> None:
    """OpenDI Model Hub CLI."""
    if ctx.invoked_subcommand is None:
        typer.echo("OpenDI CLI — use --help to see available commands.", err=True)
        raise typer.Exit(0)


@app.command()
def login() -> None:
    """Log in to the OpenDI hub."""
    typer.echo("Login not yet implemented.")


@app.command()
def logout() -> None:
    """Log out from the OpenDI hub."""
    typer.echo("Logout not yet implemented.")


@app.command()
def pull(
    name: str = typer.Argument(..., help="Model or resource name to pull"),
) -> None:
    """Pull a model or resource from the hub."""
    typer.echo(f"Pull not yet implemented for: {name}")


@app.command()
def push(
    path: str = typer.Argument(..., help="Local path to push"),
) -> None:
    """Push a model or resource to the hub."""
    typer.echo(f"Push not yet implemented for: {path}")


if __name__ == "__main__":
    app()
