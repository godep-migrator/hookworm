# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8

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
    log.info { "Configured!  Wrote config to #{cfg_file}" }
  end

  def cfg
    @cfg ||= JSON.parse(File.read(cfg_file), symbolize_names: true)
  end

  def cfg_file
    File.join(Dir.pwd, "#{File.basename($PROGRAM_NAME)}.cfg.json")
  end

  def handle(type)
    send(:"handle_#{type}")
  end

  def log
    @log ||= build_log
  end

  def input_stream
    $stdin.set_encoding('UTF-8')
  end

  def output_stream
    $stdout.set_encoding('UTF-8')
  end

  def log_stream
    $stderr.set_encoding('UTF-8')
  end

  def build_log
    logger = Logger.new(log_stream)
    logger.level = cfg[:debug] ? Logger::DEBUG : Logger::INFO
    log_level = cfg[:log_level]
    if log_level && Logger.const_defined?(log_level.upcase)
      logger.level = Logger.const_get(log_level.upcase)
    end
    logger
  end
end
