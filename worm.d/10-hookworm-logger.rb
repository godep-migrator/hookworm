#!/usr/bin/env ruby
# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8
#+ #### Hookworm Logger
#+
#+ The logger is responsible for logging valid incoming requests, optionally
#+ logging to syslog if the `syslog=true` postfix option is provided.  Log
#+ verbosity is higher if the `-d` debug flag is passed.

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

    log_payload(payload, re_serialized_payload)
    return 0
  rescue => e
    log.error { "#{e.class.name} #{e.message}" }
    log.debug { e.backtrace.join("\n") }
    return 1
  end

  def handle_travis
    payload = JSON.parse(input_stream.read, symbolize_names: true)
    re_serialized_payload = JSON.pretty_generate(payload)

    log_payload(payload, re_serialized_payload)
    return 0
  rescue => e
    log.error { "#{e.class.name} #{e.message}" }
    log.debug { e.backtrace.join("\n") }
    return 1
  end

  def log_payload(payload, re_serialized_payload)
    log.debug { "payload=#{payload.inspect}" }
    log.debug { "payload json=#{re_serialized_payload.inspect}" }

    if (cfg[:worm_flags] || {})[:syslog]
      Syslog.open($PROGRAM_NAME, Syslog::LOG_PID | Syslog::LOG_CONS) do |sl|
        sl.info(re_serialized_payload)
      end
    end

    output_stream.puts(re_serialized_payload)
  end
end

exit HookwormLogger.new.run!(ARGV) if $PROGRAM_NAME == __FILE__
