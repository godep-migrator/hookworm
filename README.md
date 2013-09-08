hookworm
========

GitHub & Travis hook receiving thingydoo.

[![Build Status](https://travis-ci.org/modcloth-labs/hookworm.png?branch=master)](https://travis-ci.org/modcloth-labs/hookworm)

## Usage

```
Usage: hookworm-server [options] [key=value...]
  -D="": Working directory (scratch pad)
  -P="": PID file (only written if flag given)
  -T=30: Timeout for handler executables (in seconds)
  -W="": Worm directory that contains handler executables
  -a=":9988": Server address
  -d=false: Show debug output
  -github.path="/github": Path to handle Github payloads
  -rev=false: Print revision and exit
  -travis.path="/travis": Path to handle Travis payloads
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

- has one of the following file extensions: `.js`, `.pl`, `.py`, `.rb`, `.sh`
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

### Included handlers

Hookworm ships with the following handlers:

#### Hookworm Annotator

The annotator is responsible for adding fields to the incoming payloads so
that subsequent handlers do not have to duplicate decision-making logic.

##### GitHub payload annotation
GitHub payloads are given the following additional fields dependending on
the presence of certain options.

###### `is_pr_merge`
Is the payload the result of a pull request merge?

###### `is_watched_branch`
Is the payload for a branch that is "watched", depending on the presence of
the `watched_branches` postfix keyword argument?

###### `has_watched_path`
Does the payload contain changes to a "watched" path, depending on the
presence of the `watched_paths` postfix keyword argument?


#### Hookworm Logger

The logger is responsible for logging valid incoming requests, optionally
logging to syslog if the `syslog=true` postfix option is provided.  Log
verbosity is higher if the `-d` debug flag is passed.


#### Hookworm Rogue Commit Handler

The rogue commit handler is specific to GitHub payloads.  It will inspect
a payload in the context of the given `watched_branches` and `watched_paths`
and send a "rogue commit email" to the email recipients given in
`email_recipients` to provide visibility roughly equivalent to those commits
that result from pull request merges.

Because the rogue commit handler is affected by so many arguments, here they
are again with more details about their associated behavior:

##### `watched_branches`
The `watched_branches` argument should be a comma-delimited list of regular
expressions, e.g.: `watched_branches='^master$,^release_[0-9]'`.  If a
commit payload is received that was not the result of a pull request merge
and the Hookworm Annotator handler has determined that the branch name
matches any of the entries in `watched_branches`, then a rogue commit email
will be sent.

##### `watched_paths`
The `watched_paths` argument should be a comma-delimited list of regular
expressions, e.g.: `watched_paths='.*\.(go|rb|py)$,bin/.*'`.  If a commit
payload is received that was not the result of a pull request merge and the
Hookworm Annotator handler has determined that one of the commits in the
payload contains a path matching any of the entries in `watched_paths`, then
a rogue commit email will be sent.

##### `email_from_addr`
The `email_from_addr` is the email address used as the `From` header and
SMTP MAIL address when sending rogue commit emails, e.g.:
`email_from_addr='hookworm-noreply@company.example.com'`.

##### `email_recipients`
The `email_recipients` argument should be a comma-delimited list of email
addresses (without display name) used in the `To` header and SMTP RCPT
addresses when sending rogue commit emails, e.g.:
`email_recipients='devs@example.com,proj+hookworm@partner.example.net'`

##### `email_uri`
The `email_uri` argument should be a well-formed URI containing the SMTP
hostname and port and potentially the username and password used for plain
SMTP auth, e.g.:
`email_uri='smtp://hookworm:secret@mailhost.example.com:1587'`




### Handler logging

Each handler that uses the `.hookworm_base.rb` has a log that writes to
`$stderr`, the level for which may be set via the `log_level` postfix
argument as long as it is a valid string log level, e.g.
`log_level=debug`.
