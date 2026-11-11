# appsvc host deploy files

Copies of the files that build and run fulfillmentd on the appsvc host (user `appservices`). The host runs
its own copies; after changing one here, copy it to the host path below.

| In this directory | On the host |
| --- | --- |
| `build-fulfillmentd-container.sh`, `restart-fullfillmentd.sh` | `~/` |
| `fulfillmentd.dockerfile` | `~/src/` |
| `fulfillmentd.service` | `~/.config/systemd/user/` (then `systemctl --user daemon-reload`) |

The build clones this repo's branch (default `main`) from GitHub, so changes must be pushed before building.
The Dockerfile also copies `~/src/configs/fulfillmentd-test.toml` from the host; it is not in this repo.

## Release

```sh
~/build-fulfillmentd-container.sh [branch]
~/restart-fullfillmentd.sh
curl -s localhost:2023/version; echo; curl -s localhost:2023/health; echo
```
