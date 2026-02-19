Fully AI generated

# Setup
Copy `config.yaml.dist` to `config.yaml` and adjust the tags to watch and add an absolute path to the `state.json` file.

Run `make run` to build the binary and fetch inital image update timestamps.

There is a script to raise Linux notifications in case image updates are found.
This script would usually be executed via crontab like

```
CRON_TZ=Europe/Berlin
30 9  * * 1-5 /absolute/path/to/registry-ping/notify-run.sh
30 13 * * 1-5 /absolute/path/to/registry-ping/notify-run.sh
```
