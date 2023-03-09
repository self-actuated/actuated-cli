# actuated-cli

## Installation

Download the latest release from the [releases page](https://github.com/self-actuated/actuated-cli/releases).

Then add the URL to either .bashrc or .zshrc:

```bash
export ACTUATED_API=https://example.com
```

Or, run this command in a shell before executing any of the CLI commands.

## View queued jobs

```bash
actuated-cli jobs \
    --pat ~/reader.txt \
    --owner actuated-samples
```

## View runners for organization

```bash
actuated-cli runners \
    --pat ~/reader.txt \
    --owner actuated-samples
```


## Check the logs of VMs

View the serial console and systemd output of the VMs launched on a specific server.

* Check for timeouts with GitHub's control-plane
* View output from the GitHub runner binary
* See boot-up messages
* Check for errors if the GitHub Runner binary is out of date

```bash
actuated-cli logs \
    --pat ~/reader.txt \
    --host runner1 \
    --owner actuated-samples \
    --age 15m
```

The age is specified as a Go duration i.e. `60m` or `24h`.

## Check the logs of the actuated agent service

Show the logs of the actuated agent binary running on your server.

View VM launch times, etc.

```bash
actuated-cli agent-logs \
    --pat ~/reader.txt \
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
    --pat ~/reader.txt \
    --owner actuated-samples
```

## JSON mode

Add `--json` to any command to get JSON output for scripting.

API rate limits apply, so do not run the CLI within a loop or `watch` command.

## Staff mode

The `--staff` flag can be added to the `runners`, `jobs` and the `repair` commands by OpenFaaS Ltd staff to support actuated customers.

## Help & support

Reach out to our team on Slack.
