require 'json'
require 'logger'

module HookwormBase
  def run!(argv)
    action = argv.first
    if %(configure handle).include?(action)
      return send(*argv)
    else
      abort("I don't know how to #{action.inspect}")
    end
  end

  private

  def configure
    @cfg = JSON.parse(input_stream.read, symbolize_names: true)
    File.open(cfg_file, 'w') do |f|
      f.puts JSON.pretty_generate(@cfg)
    end
    log.info "Configured!  Wrote config to #{cfg_file}"
  end

  def cfg
    @cfg ||= JSON.parse(File.read(cfg_file), symbolize_names: true)
  end

  def cfg_file
    File.join(Dir.pwd, "#{File.basename($0)}.cfg.json")
  end

  def handle(type)
    send(:"handle_#{type}")
  end

  def log
    @log ||= Logger.new(log_stream)
  end

  def input_stream
    ($hookworm_stdin || $stdin).set_encoding('UTF-8')
  end

  def output_stream
    ($hookworm_stdout || $stdout).set_encoding('UTF-8')
  end

  def log_stream
    ($hookworm_stderr || $stderr).set_encoding('UTF-8')
  end
end
