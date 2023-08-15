# actuated-cli

The actuated-cli requires an access token for GitHub and is designed to be used by a user with access to the actuated dashboard.

Most of the operations on the actuated dashboard are available as CLI commands.

## Installation

Download the latest release from the [releases page](https://github.com/self-actuated/actuated-cli/releases).

Then add the URL to either .bashrc or .zshrc:

```bash
export ACTUATED_URL=https://example.com
```

Or, run this command in a shell before executing any of the CLI commands.

## Obtain a token from GitHub for your own account

You can perform this step using GitHub's Device Flow:

```bash
actuated-cli auth
```

Or you can obtain a Personal Access Token (PAT) manually from [https://github.com/settings/tokens](https://github.com/settings/tokens)

In either case, saving the token to `$HOME/.actuated/PAT` will mean you can avoid having to pass in the `--token` flag to each command.

## View queued jobs

```bash
actuated-cli jobs \
    --owner actuated-samples
```

## View runners for organization

```bash
actuated-cli runners \
    --owner actuated-samples
```

## View SSH sessions available:

```bash
actuated-cli ssh \
    ssh ls
```

Hosts are ordered by the connected time.

```bash
| NO  |   ACTOR   |                 HOSTNAME                 |  RX   |  TX   | CONNECTED |
|-----|-----------|------------------------------------------|-------|-------|-----------|
|   1 | alexellis | 6aafd53144e2f00ef5cd2c16681eeab4712561a6 | 13679 | 10371 | 6m4s      |
|   2 | alexellis | fv-az268-245                             | 23124 | 13828 | 12m2s     |
```

## Connect to an SSH session

Connect to the first available session from your account:

```bash
actuated-cli ssh \
    ssh connect
```

Connected to the second session in the list:

```bash
actuated-cli ssh \
    ssh connect 2
```

Connect to a specific session by hostname:

```bash
actuated-cli ssh \
    ssh connect runner1
```

Connect to a specific session with a host prefix:

```bash
actuated-cli ssh \
    ssh connect 6aafd
```

## Check the logs of VMs

View the serial console and systemd output of the VMs launched on a specific server.

* Check for timeouts with GitHub's control-plane
* View output from the GitHub runner binary
* See boot-up messages
* Check for errors if the GitHub Runner binary is out of date

```bash
actuated-cli logs \
    --host runner1 \
    --owner actuated-samples \
    --age 15m
```

The age is specified as a Go duration i.e. `60m` or `24h`.

You can also get the logs for a specific runner by using the `--id` flag.

```bash
actuated-cli logs \
    --host runner1 \
    --owner actuated-samples \
    --id ea5c285282620927689d90af3cfa3be2d5e2d004
```

## Check the logs of the actuated agent service

Show the logs of the actuated agent binary running on your server.

View VM launch times, etc.

```bash
actuated-cli agent-logs \
    --host runner1 \
    --owner actuated-samples \
    --age 15m
```

## Schedule a repair to re-queue jobs

If a job has been retried for 30 minutes, without a runner to take it, it'll be taken off the queue.

This command will re-queue all jobs that are in a "queued" state.

Run with sparingly because it will launch one VM per job queued.

```bash
actuated-cli repair \
    --owner actuated-samples
```

## Rescue a remote server

Restart the agent by sending a `kill -9` signal:

```bash
actuated-cli restart \
    --owner actuated-samples \
    --host runner1
```

Any inflight VMs will be killed, see also: `actuated-cli update --force`

Reboot the machine, if in an unrecoverable position:

```bash
actuated-cli restart \
    --owner actuated-samples \
    --host runner1 \
    --reboot
```

Use with caution, since this may not perform a safe and clean shutdown.

## JSON mode

Add `--json` to any command to get JSON output for scripting.

API rate limits apply, so do not run the CLI within a loop or `watch` command.

## Staff mode

The `--staff` flag can be added to the `runners`, `jobs` and the `repair` commands by OpenFaaS Ltd staff to support actuated customers.

## Help & support

Reach out to our team on Slack.
