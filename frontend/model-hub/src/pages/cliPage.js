//
// COPYRIGHT OpenDI
//

import React from 'react';
import {
    Box,
    Button,
    Card,
    Typography,
    Container,
    Divider
} from "@mui/material";
import CodeIcon from '@mui/icons-material/Code';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';

const CopyableCommandBlock = ({ command, multiline = false }) => {
    const [copied, setCopied] = React.useState(false);

    const handleCopy = async () => {
        try {
            await navigator.clipboard.writeText(command);
            setCopied(true);
            setTimeout(() => setCopied(false), 1400);
        } catch (_) {
            // no-op fallback; keep UI simple
        }
    };

    return (
        <Box
            sx={{
                bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5',
                color: 'text.primary',
                p: 2,
                borderRadius: 1,
                mb: 2,
                fontFamily: 'monospace',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: multiline ? 'flex-start' : 'center',
                gap: 2,
            }}
        >
            <Box component="pre" sx={{ m: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word', flex: 1 }}>
                {command}
            </Box>
            <Button
                size="small"
                variant="outlined"
                startIcon={<ContentCopyIcon fontSize="small" />}
                onClick={handleCopy}
            >
                {copied ? 'Copied' : 'Copy'}
            </Button>
        </Box>
    );
};

const TOP_LEVEL_HELP_OUTPUT = `Usage: opendi [OPTIONS] COMMAND [ARGS]...

OpenDI Model Hub CLI for discovering and managing CDM models.

Options:
  --install-completion     Install completion for the current shell.
  --show-completion        Show completion for the current shell.
  --help                   Show this message and exit.

Authentication:
  login     Log in to the OpenDI hub (browser-assisted device flow).
  whoami    Show the currently logged-in account.
  logout    Log out and clear stored credentials.

Resource Commands:
  inspect   Show repository metadata or tag metadata.
  create    Create resources (create repo).
  delete    Delete resources (delete repo/tag/local).
  list      List resources (list repos/local).
  add       Add resources (add tag).

Discovery:
  search    Search public repositories on the hub.

Model Operations:
  pull      Pull a model from the hub into local cache.
  push      Push a local CDM JSON file to the hub.
  save      Save a model from the hub to a local JSON file.
  diff      Compare two CDM models.
  validate  Validate a local CDM JSON file against the schema.`;

const CliDownloadPage = () => {
    return (
        <Container maxWidth="md" sx={{ py: 4 }}>
            <Card sx={{ p: 4 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                    <CodeIcon sx={{ fontSize: 40, mr: 2, color: 'primary.main' }} />
                    <Typography variant="h4" fontWeight="bold">
                        OpenDI CLI
                    </Typography>
                </Box>
                
                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                    Command-line interface for the OpenDI Model Hub. Manage and interact with model repositories from your terminal.
                </Typography>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Installation
                </Typography>

                <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
                    Requirements
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                    Python 3.10 or higher.
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                    Download Python from{' '}
                    <a
                        href="https://www.python.org/downloads/"
                        target="_blank"
                        rel="noopener noreferrer"
                        style={{ color: 'inherit', textDecoration: 'underline' }}
                    >
                        python.org/downloads
                    </a>
                    . On Windows, make sure Python is added to <code style={{ fontFamily: 'monospace' }}>PATH</code>
                    {' '}during installation.
                </Typography>

                <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
                    Install with pipx (Windows)
                </Typography>
                <CopyableCommandBlock
                    multiline
                    command={`py -m pip install --user pipx
py -m pipx install opendi`}
                />

                <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
                    Install with pipx (macOS / Linux)
                </Typography>
                <CopyableCommandBlock
                    multiline
                    command={`python3 -m pip install --user pipx
python3 -m pipx install opendi`}
                />

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Getting Started
                </Typography>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    1. Verify Installation
                </Typography>
                <CopyableCommandBlock command="opendi --help" />

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    2. Log In
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Authenticate with your OpenDI account. This will open your browser to complete login:
                </Typography>
                <CopyableCommandBlock command="opendi login" />
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Your credentials are securely stored in your OS credential manager (Windows Credential Manager, macOS Keychain, or Linux Secret Service).
                </Typography>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    3. Verify Login
                </Typography>
                <CopyableCommandBlock command="opendi whoami" />

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    4. Available Commands
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Snapshot of <code style={{ fontFamily: 'monospace' }}>opendi --help</code>:
                </Typography>
                <Box
                    component="pre"
                    sx={{
                        bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5',
                        color: 'text.primary',
                        p: 2,
                        borderRadius: 1,
                        mb: 2,
                        fontFamily: 'monospace',
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-word',
                        m: 0,
                    }}
                >
                    {TOP_LEVEL_HELP_OUTPUT}
                </Box>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Help & Documentation
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    Get help on any command:
                </Typography>
                <Box
                    sx={{
                        bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5',
                        color: 'text.primary',
                        p: 2,
                        borderRadius: 1,
                        mb: 2,
                        fontFamily: 'monospace',
                    }}
                >
                    opendi [command] --help
                </Box>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Uninstall
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    Windows:
                </Typography>
                <CopyableCommandBlock command="py -m pipx uninstall opendi" />
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    macOS / Linux:
                </Typography>
                <CopyableCommandBlock command="python3 -m pipx uninstall opendi" />
            </Card>
        </Container>
    );
};

export default CliDownloadPage;