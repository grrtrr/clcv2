## Commandline console for CenturyLink Cloud operations

This is `clconsole`, a utility to manage CenturyLink cloud VMs and hardware groups.
It is aimed for daily use and so focuses on grouping the most-used commands in a straightforward way.

```
Usage:
  clconsole [command]

Available Commands:
  archive         Archive server(s)
  restore         Restore server/group from archive
  cpu             Set server #CPU
  mem             Set server memory
  desc            Change server description
  pass            Set or generate server password
  clone           Clone existing server
  create          Create server from template/source
  creds           Print login credentials of server(s)
  rm              Delete server(s)/group(s) (CAUTION)
  mkdir           Create a new folder
  mv              Move server(s)/group(s) into different folder
  nets            Show available networks
  nic             Manage server NICs
  off             Power-off or suspend server(s)
  ip              Add a public IP to a server
  rawdisk         Add storage to a server
  rename          Rename group
  restart         Reboot or reset server(s)
  on              Power on server(s)
  snapshot        Snapshot server(s)
  delsnap         Delete snapshot of server(s)
  revert          Revert server(s) to snapshot
  ls              Show server(s)/groups(s)
  templates       List available templates
  wait            Await completion of queue job and report status
```

## Building

By default, `make` will generate the executable for Linux.

### Windows Version

To build for _Windows_ (`clconsole.exe`), type
```bash
> make windows
```

### Mac Version

To build for _Mac_ (`clconsole.mac`), type
```bash
> make mac
```

### CenturyLink Cloud Login

At first login you will need credentials (username, password, and optionally account).

If you already have an existing installation of [clc-go-cli](https://github.com/CenturyLinkCloud/clc-go-cli),
your credentials will be imported.

Otherwise, you can use the following _options_ and _environment variables_  (which take preference over options):
- `-u/--username` or `$CLC_USER`,
- `-p/--password` or `$CLC_PASSWORD`,
- `-a/--account` or `$CLC_ACCOUNT` (if you use a sub-account).

Once supplied, credentials are stored in `$HOME/.clc/client_config.yaml` on Linux/Mac and `C:\Users\%username%\clc\client_config.yml`
on Windows.

To speed up login, the program also _reuses the bearer token_ (which is valid for up to 2 weeks), by storing it in the same folder
under `credentials.json`. Should this token expire, the library will re-login to retrieve (and then save) a new token.

You can also set a _default data centre location_ via  `-l/--location` or `$CLC_LOCATION`. The program will remember the
last datacentre, which is handy when doing multiple operations in the same location.

### Bash Auto-Completion

This program has support for [bash completion](https://www.gnu.org/software/bash/manual/html_node/Programmable-Completion.html),
which you can simply _generate_ by running
```bash
> clconsole bash-completion [-f /path/to/completion.file]
```
This will generate the bash-completion file. If that file is in a _non-standard location_, run
```bash
> source /path/to/completion.file
```
to activate the changes.

If no `-f` flag is provided, the _default location_ is `/etc/bash_completion.d/clconsole.sh` (requires running with `root` permissions). For more details, run
```bash
> clconsole bash-completion -h          # or --help
```

## Credits

This tool has benefited a lot from studying the code of [clc-go-cli](https://github.com/CenturyLinkCloud/clc-go-cli),
whose authors have done a great job.
