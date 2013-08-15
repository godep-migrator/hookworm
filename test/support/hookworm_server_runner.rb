require_relative 'annunciation'

class HookwormServerRunner
  include Annunciation

  attr_reader :server_binary, :addr, :port, :dir, :start_time, :pidfile
  attr_reader :startup_sleep, :server_pid, :options, :logfile

  def initialize(options = {})
    @options = options
    @start_time = options.delete(:start) || Time.now.utc
    start = start_time.strftime('%Y%m%d%H%M%S')
    @addr = options.fetch('-a')
    @port = Integer(addr.gsub(/.*:/, ''))
    @server_binary = "#{ENV['GOPATH'].split(/:/).first}/bin/hookworm-server"
    @logfile = File.expand_path(
      "../../../log/hookworm-server-#{port}-#{start}.log",
      __FILE__
    )
    @pidfile = options['-P'] || "hookworm-server-#{port}.pid"
    @dir = options['-D']
    @startup_sleep = Float(ENV['HOOKWORM_STARTUP_SLEEP'] || 0.5)

    if !File.exist?(server_binary)
      raise "Can't locate `hookworm-server` binary! " <<
            "(it's not here: #{server_binary.inspect})"
    end
  end

  def start
    if dir
      FileUtils.mkdir_p(dir)
    end
    process_command = command
    announce! "Starting #{description}"
    announce! "  ---> #{process_command}"
    @server_pid = Process.spawn(process_command)
    sleep startup_sleep
    server_pid
  end

  def stop
    real_pid = Integer(File.read(pidfile).chomp) rescue nil
    if server_pid && real_pid
      announce! "Stopping #{description} " <<
                "(shell PID=#{server_pid}, server PID=#{real_pid})"

      [real_pid, server_pid].each do |pid|
        Process.kill(:TERM, pid) rescue nil
      end
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
      cmd << k << "'#{v}'"
    end
    cmd << ">> #{logfile} 2>&1"
    cmd.join(' ')
  end

  def dump_log
    announce! "Dumping #{logfile}"
    File.read(logfile).split($/).each do |line|
      announce! "--> #{line}"
    end
  end
end
