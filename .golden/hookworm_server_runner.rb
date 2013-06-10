require_relative 'server_runner_methods'

class HookwormServerRunner
  include ServerRunnerMethods

  def initialize(options = {})
    @options = options
    @start_time = options.delete(:start) || Time.now.utc
    start = start_time.strftime('%Y%m%d%H%M%S')
    @addr = options.fetch('-a')
    @port = Integer(addr.gsub(/.*:/, ''))
    @server_binary = "#{ENV['GOPATH'].split(/:/).first}/bin/hookworm-server"
    @logfile = File.expand_path(
      "../../log/hookworm-server-#{start}-#{port}.log",
      __FILE__
    )
    @pidfile = options['-P'] || "hookworm-server-#{port}.pid"
    @startup_sleep = Float(ENV['HOOKWORM_STARTUP_SLEEP'] || 0.5)

    if !File.exist?(server_binary)
      raise "Can't locate `hookworm-server` binary! " <<
            "(it's not here: #{server_binary.inspect})"
    end
  end

  def description
    "hookworm server with address #{addr}"
  end

  def command
    cmd = [server_binary]
    options.each do |k,v|
      if v.nil?
        cmd << k
        next
      end
      cmd << k << v
    end
    cmd << ">> #{logfile} 2>&1"
    cmd.join(' ')
  end
end
