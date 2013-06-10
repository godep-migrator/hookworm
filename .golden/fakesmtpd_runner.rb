require_relative 'server_runner_methods'

class FakeSMTPdRunner
  include ServerRunnerMethods

  attr_reader :http_port

  def initialize(options = {})
    @start_time = Time.now.utc
    @start = @start_time.strftime('%Y%m%d%H%M%S')
    @port = Integer(options.fetch(:port))
    @http_port = port + 1
    @dir = options.fetch(:dir)
    @pidfile = options[:pidfile] || 'fakesmtpd.pid'
    @logfile = options[:logfile] ||
      File.expand_path("../../log/fakesmtpd-#{@start}.log", __FILE__)
    @startup_sleep = options[:startup_sleep] || 0.5
  end

  def description
    "fakesmtpd server on port #{port}"
  end

  def command
    [
      File.expand_path('../../fakesmtpd', __FILE__),
      port.to_s,
      dir,
      pidfile,
    ].join(' ') << " >> #{logfile} 2>&1"
  end
end
