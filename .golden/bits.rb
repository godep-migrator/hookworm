require 'net/http'

GOLD = "\033\[33;1m"
RESET = "\033\[0m"
GREEN = "\033\[32m"
RED = "\033\[31m"
BRIGHT_GREEN = "\033\[32;1m"
BRIGHT_RED = "\033\[31;1m"

def announce!(something)
  $stderr.puts "#{GOLD}golden#{RESET}: #{GREEN}#{something}#{RESET}"
end

class MiniTestReporter
  def puts(*args)
    args.each { |arg| announce! arg }
  end

  alias print puts
end
