hookworm
========

GitHub & Travis hook receiving thingydoo.

[![Build Status](https://travis-ci.org/modcloth-labs/hookworm.png?branch=master)](https://travis-ci.org/modcloth-labs/hookworm)

## Usage

```
Usage: hookworm-server [options] [key=value...]
  -D="": Working directory (scratch pad) [HOOKWORM_WORKING_DIR]
  -P="": PID file (only written if flag given) [HOOKWORM_PID_FILE]
  -S="": Public static directory (default $PWD/public) [HOOKWORM_STATIC_DIR]
  -T=30: Timeout for handler executables (in seconds) [HOOKWORM_HANDLER_TIMEOUT]
  -W="": Worm directory that contains handler executables [HOOKWORM_WORM_DIR]
  -a=":9988": Server address [HOOKWORM_ADDR]
  -d=false: Show debug output [HOOKWORM_DEBUG]
  -github.path="/github": Path to handle Github payloads [HOOKWORM_GITHUB_PATH]
  -rev=false: Print revision and exit
  -travis.path="/travis": Path to handle Travis payloads [HOOKWORM_TRAVIS_PATH]
  -version=false: Print version and exit
  -version+=false: Print version, revision, and build tags
```

Hookworm is designed to listen for GitHub and Travis webhook payloads
and delegate handling to a pipeline of executables.  In this way, the
long-running server process stays smallish (~6MB) and any increase in
memory usage at payload handling time is ephemeral, assuming the handler
executables aren't doing anything silly.

An example invocation that uses the handler executables shipped with
hookworm would look like this, assuming the hookworm repo has been
cloned into `/var/lib/hookworm`:

``` bash
mkdir -p /var/run/hookworm-main
hookworm-server -d \
  -D /var/run/hookworm-main \
  -W /var/lib/hookworm/worm.d \
  syslog=yes >> /var/log/hookworm-main.log 2>&1
```

### Handler contract

Handler executables are expected to fulfill the following contract:

- has one of the following file extensions: `.js`, `.pl`, `.py`, `.rb`, `.sh`, `.bash`
- does not begin with `.` (hidden file)
- accepts a positional argument of `configure`
- accepts positional arguments of `handle github`
- accepts positional arguments of `handle travis`
- writes only the (potentially modified) payload to standard output
- exits `0` on success
- exits `78` on no-op (roughly `ENOSYS`)

It is up to the handler executable to decide what is done for each
command invocation.  The execution environment includes the
`HOOKWORM_WORKING_DIR` variable, which may be used as a scratch pad for
temporary files.

#### `<interpreter> <handler-executable> configure`

The `configure` command is invoked at server startup time for each
handler executable, passing the handler configuration as a JSON object
on the standard input stream.  The configuration object is guaranteed to
have all of the values provided as flags to `hookworm-server`.

Additionally, any key-value pairs provided as postfix arguments will be
added to a `worm_flags` hash such as the `syslog=yes` argument given in
the above example.  Bare keys are assigned a JSON value of `true`.
String values of `true`, `yes`, and `on` are converted to JSON `true`,
and string values of `false`, `no`, and `off` are converted to JSON
`false`.

#### `<interpreter> <handler-executable> handle github`

The `handle github` command is invoked whenever a payload is received at
the GitHub-handling path (`/github` by default).  The payload is passed
to the handler executable as a JSON object on the standard input stream.

#### `<interpreter> <handler-executable> handle travis`

The `handle travis` command is invoked whenever a payload is received at
the Travis-handling path (`/travis` by default).  The payload is passed
to the handler executable as a JSON object on the standard input stream.

### Handler logging

Each handler that uses the `hookworm-base` gem has a log that writes to
`$stderr`, the level for which may be set via the `log_level` postfix
argument as long as it is a valid string log level, e.g.
`log_level=debug`.
