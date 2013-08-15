hookworm
========

github hook receiving thingydoo

```
Usage: hookworm-server [options] [key=value...]
  -D="": Working directory (scratch pad)
  -P="": PID file (only written if flag given)
  -T=30: Timeout for handler executables (in seconds)
  -W="": Worm directory that contains handler executables
  -a=":9988": Server address
  -b="": Watched branches (comma-delimited regexes)
  -d=false: Show debug output
  -e="smtp://localhost:25": Email server address
  -f="hookworm@localhost": Email from address
  -github.path="/github": Path to handle Github payloads
  -p="": Watched paths (comma-delimited regexes)
  -r="": Email recipients (comma-delimited)
  -rev=false: Print revision and exit
  -travis.path="/travis": Path to handle Travis payloads
  -version=false: Print version and exit
  -version+=false: Print version, revision, and build tags
```
