#!/usr/bin/env ruby

require 'json'
require 'logger'
require 'syslog'

class HookwormLogger
  def run!(argv)
    action = argv.first
    if %(configure handle).include?(action)
      send(*argv)
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
    if type != 'github'
      abort("Unknown payload type #{type.inspect}")
    end

    payload = JSON.parse(input_stream.read, symbolize_names: true)
    re_serialized_payload = JSON.pretty_generate(payload)

    log.info "Pull request merge? #{payload[:is_pr_merge]}"
    if cfg[:debug]
      log.info "payload=#{payload.inspect}"
      log.info "payload json=#{re_serialized_payload.inspect}"
    end

    if cfg[:syslog]
      Syslog.open($0, Syslog::LOG_PID | Syslog::LOG_CONS) do |syslog|
        syslog.info(re_serialized_payload)
      end
    end

    return 0
  rescue => e
    log.error "#{e.class.name} #{e.message}"
    if cfg[:debug]
      log.error e.backtrace.join("\n")
    end
    return 1
  end

  def log
    @log ||= Logger.new(log_stream)
  end

  def input_stream
    $hookworm_stdin || $stdin
  end

  def log_stream
    $hookworm_stderr || $stderr
  end
end

if $0 == __FILE__
  HookwormLogger.new.run!(ARGV)
end
