#!/usr/bin/env ruby

require 'json'
require 'logger'
require 'syslog'

class EventLogHandler
  def run!(*argv)
    action = argv.first
    if %(configure handle).include?(action)
      send(*argv)
    else
      raise NoMethodError.new(action)
    end
  end

  private

  def configure
    @cfg = JSON.parse($stdin.read, symbolize_names: true)
    File.open(cfg_file, 'w') do |f|
      f.puts JSON.pretty_generate(@cfg)
    end
  end

  def cfg
    @cfg ||= JSON.parse(File.read(cfg_file), symbolize_names: true)
  end

  def cfg_file
    if cfg[:working_dir]
      File.join(cfg[:working_dir], "#{File.basename($0)}.cfg.json")
    else
      $stderr.puts "WARNING: no working dir set in config, so using $PWD"
      File.join(Dir.pwd, "#{File.basename($0)}.cfg.json")
    end
  end

  def handle(type = 'github')
    if type != 'github'
      raise RuntimeError("Unknown payload type #{type.inspect}")
    end

    payload = JSON.parse($stdin.read, symbolize_names: true)
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
    @log ||= Logger.new($stderr)
  end
end

if $0 == __FILE__
  EventLogHandler.new.run!(ARGV.first)
end
