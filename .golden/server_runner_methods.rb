require_relative 'bits'
require 'fileutils'

module ServerRunnerMethods
  include Bits

  attr_reader :server_binary, :addr, :port, :dir, :start_time, :pidfile
  attr_reader :startup_sleep, :server_pid, :options, :logfile

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
    "???"
  end

  def command
    'echo foo'
  end

  def dump_log
    announce! "Dumping #{logfile}"
    File.read(logfile).split($/).each do |line|
      announce! "--> #{line}"
    end
  end
end
