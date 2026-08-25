from __future__ import annotations

import sys
from unittest.mock import MagicMock, patch

from memlore.adapters.cli.main import _warn_go_preferred, main


def test_warn_go_preferred_mentions_legacy_and_command(capsys) -> None:
    _warn_go_preferred("mcp")
    err = capsys.readouterr().err
    assert "legacy" in err.lower()
    assert "mcp" in err


def test_serve_prints_deprecation_before_uvicorn(capsys) -> None:
    with patch("uvicorn.run") as run:
        with patch.object(sys, "argv", ["memlore", "serve"]):
            main()
        run.assert_called_once()
    assert "legacy" in capsys.readouterr().err.lower()


def test_mcp_prints_deprecation_before_server(capsys) -> None:
    mock_server = MagicMock()
    with patch("memlore.adapters.cli.main.create_mcp_server", return_value=mock_server):
        with patch.object(sys, "argv", ["memlore", "mcp"]):
            main()
        mock_server.run.assert_called_once_with(transport="stdio")
    assert "legacy" in capsys.readouterr().err.lower()
