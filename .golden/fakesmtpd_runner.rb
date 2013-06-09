require_relative 'server_runner_methods'

class FakeSMTPdRunner
  include ServerRunnerMethods

  def initialize(options = {})
    @start_time = Time.now.utc
    @start = @start_time.strftime('%Y%m%d%H%M%S')
    @port = options.fetch(:port)
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
