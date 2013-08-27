#!/usr/bin/env ruby
#+ #### Hookworm Logger
#+
#+ The logger is responsible for logging valid incoming requests, optionally
#+ logging to syslog if the `syslog=true` postfix option is provided.  Log
#+ verbosity is higher if the `-d` debug flag is passed.
#+

require 'json'
require 'logger'
require 'syslog'

require_relative '.hookworm_base'

class HookwormLogger
  include HookwormBase

  private

  def handle_github
    payload = JSON.parse(input_stream.read, symbolize_names: true)
    re_serialized_payload = JSON.pretty_generate(payload)

    log.info "Pull request merge? #{payload[:is_pr_merge]}"
    if cfg[:debug]
      log.info "payload=#{payload.inspect}"
      log.info "payload json=#{re_serialized_payload.inspect}"
    end

    if (cfg[:worm_flags] || {})[:syslog]
      Syslog.open($0, Syslog::LOG_PID | Syslog::LOG_CONS) do |syslog|
        syslog.info(re_serialized_payload)
      end
    end

    output_stream.puts(re_serialized_payload)

    return 0
  rescue => e
    log.error "#{e.class.name} #{e.message}"
    if cfg[:debug]
      log.error e.backtrace.join("\n")
    end
    return 1
  end

  def handle_travis
    payload = JSON.parse(input_stream.read, symbolize_names: true)
    re_serialized_payload = JSON.pretty_generate(payload)

    if cfg[:debug]
      log.info "payload=#{payload.inspect}"
      log.info "payload json=#{re_serialized_payload.inspect}"
    end

    if (cfg[:worm_flags] || {})[:syslog]
      Syslog.open($0, Syslog::LOG_PID | Syslog::LOG_CONS) do |syslog|
        syslog.info(re_serialized_payload)
      end
    end

    output_stream.puts(re_serialized_payload)

    return 0
  rescue => e
    log.error "#{e.class.name} #{e.message}"
    if cfg[:debug]
      log.error e.backtrace.join("\n")
    end
    return 1
  end
end

if $0 == __FILE__
  HookwormLogger.new.run!(ARGV)
end
