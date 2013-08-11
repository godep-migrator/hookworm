hookworm
========

github hook receiving thingydoo

```
Usage: hookworm-server [options]
  -D="": Working directory (scratch pad)
  -P="": PID file (only written if flag given)
  -S=false: Send all received events to syslog
  -T=30: Timeout for handler executables (in seconds)
  -W="": Worm directory that contains handler executables
  -a=":9988": Server address
  -b="": Watched branches (comma-delimited regexes)
  -d=false: Show debug output
  -e="smtp://localhost:25": Email server address
  -f="hookworm@localhost": Email from address
  -p="": Watched paths (comma-delimited regexes)
  -r="": Email recipients (comma-delimited)
  -rev=false: Print revision and exit
  -version=false: Print version and exit
```
