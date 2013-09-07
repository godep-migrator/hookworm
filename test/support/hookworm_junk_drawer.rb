# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8

require 'json'
require 'net/http'
require 'open3'
require 'tmpdir'
require 'uri'

require 'mail'

require_relative 'servers'

module HookwormJunkDrawer
  include Mtbb::NetThings

  def base_command
    ['echo']
  end

  def handle(stdin_string, args)
    ENV['HOOKWORM_WORKING_DIR'] = tempdir
    command = base_command + args
    execute_command_in_tmpdir(command, stdin_string)
  end

  def execute_command_in_tmpdir(command, stdin_string)
    Dir.chdir(tempdir) { execute_command(command, stdin_string) }
  end

  def execute_command(command, stdin_string)
    out, err = '', ''
    exit_status = 1

    Open3.popen3(*command) do |stdin, stdout, stderr, wait_thr|
      stdin.write stdin_string
      stdin.close
      out << stdout.read
      err << stderr.read
      exit_status = wait_thr.value
    end

    [exit_status, out, err]
  end

  def tempdir
    @tempdir ||= setup_tempdir
  end

  def setup_tempdir
    dir = Dir.mktmpdir
    at_exit { FileUtils.rm_rf(dir) }
    dir
  end

  def github_payload(name, format)
    case format
    when :form
      "payload=#{URI.escape(github_payload_string(name))}"
    when :json
      github_payload_string(name)
    end
  end

  def github_payload_hash(name)
    JSON.parse(github_payload_string(name), symbolize_names: true)
  end

  def github_payload_string(name)
    File.read(github_payload_file(name))
  end

  def github_payload_file(name)
    File.expand_path(
      "../../../sampledata/github-payloads/#{name.to_s}.json", __FILE__
    )
  end

  def post_github_payload(port, payload_name, format)
    pre_request_messages = current_mail_messages
    response = post_request(
      port: port,
      body: github_payload(payload_name, format),
      path: '/github',
      content_type: content_type(format)
    )
    [response, current_mail_messages - pre_request_messages]
  end

  def content_type(format)
    {
      form: 'application/x-www-form-urlencoded',
      json: 'application/json'
    }[format]
  end

  def travis_payload(name)
    "payload=#{URI.escape(travis_payload_string(name))}"
  end

  def travis_payload_hash(name)
    JSON.parse(travis_payload_string(name), symbolize_names: true)
  end

  def travis_payload_string(name)
    File.read(travis_payload_file(name))
  end

  def travis_payload_file(name)
    File.expand_path(
      "../../../sampledata/travis-payloads/#{name.to_s}.json", __FILE__
    )
  end

  def post_travis_payload(port, payload_name)
    pre_request_messages = current_mail_messages
    response = post_request(
      port: port, body: travis_payload(payload_name), path: '/travis'
    )
    [response, current_mail_messages - pre_request_messages]
  end

  def current_mail_messages
    JSON.parse(
      get_request(
        path: '/messages', port: Mtbb.server(:fakesmtpd).port + 1
      ).body
    ).fetch('_embedded').fetch('messages')
      .map { |h| h['_links']['self']['href'] }
  end

  def clear_mail_messages
    delete_request(path: '/messages', port: Mtbb.server(:fakesmtpd).port + 1)
  end

  def last_message
    message = JSON.parse(get_request(
      path: current_mail_messages.last, port: Mtbb.server(:fakesmtpd).port + 1
    ).body)
    return Mail.new if message['body'].nil?
    Mail.new(message['body'].join("\n"))
  end

  def last_message_header(header_name)
    last_message[header_name].to_s
  end
end
