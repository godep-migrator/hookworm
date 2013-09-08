# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8

require 'json'
require 'logger'
require 'uri'

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

  def handle_github
    78
  end

  def handle_travis
    78
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

class HookwormEmailer
  def initialize(email_uri)
    @email_uri = URI(email_uri)
  end

  def send(from, to, msg)
    Net::SMTP.start(*smtp_args) do |smtp|
      smtp.enable_ssl if @email_uri.scheme == 'smtps'
      smtp.send_message(msg, from, to)
    end
  end

  private

  def smtp_args
    [
      @email_uri.host,
      @email_uri.port,
      'localhost',
      @email_uri.user,
      @email_uri.password,
      @email_uri.user ? :plain : nil
    ]
  end
end

class String
  def commasplit
    split(',').map(&:strip)
  end

  def cleanquotes
    gsub(/["']/, '')
  end

  def to_plaintext
    gsub(/\n/, '\n').gsub(/\t/, '\t')
  end

  def to_html
    gsub(/\n/, '<br />').gsub(/\t/, '    ')
  end
end

module URI
  unless @@schemes['SMTP']
    class SMTP < Generic
      DEFAULT_PORT = 587
    end
    @@schemes['SMTP'] = SMTP
  end

  unless @@schemes['SMTPS']
    class SMTPS < Generic
      DEFAULT_PORT = 587
    end
    @@schemes['SMTPS'] = SMTPS
  end
end
