require 'json'
require 'net/http'
require 'uri'
require 'mail'

require_relative 'servers'

module HookwormJunkDrawer
  include Mtbb::NetThings

  def github_payload(name)
    "payload=#{URI.escape(github_payload_string(name))}"
  end

  def github_payload_hash(name)
    JSON.parse(github_payload_string(name), symbolize_names: true)
  end

  def github_payload_string(name)
    File.read(github_payload_file(name))
  end

  def github_payload_file(name)
    File.expand_path("../../../sampledata/github-payloads/#{name.to_s}.json", __FILE__)
  end

  def post_github_payload(port, payload_name)
    pre_request_messages = current_mail_messages
    response = post_request(port: port, body: github_payload(payload_name), path: '/github')
    [response, current_mail_messages - pre_request_messages]
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
    File.expand_path("../../../sampledata/travis-payloads/#{name.to_s}.json", __FILE__)
  end

  def post_travis_payload(port, payload_name)
    pre_request_messages = current_mail_messages
    response = post_request(port: port, body: travis_payload(payload_name), path: '/travis')
    [response, current_mail_messages - pre_request_messages]
  end

  def current_mail_messages
    JSON.parse(
      get_request(path: '/messages', port: Mtbb.server(:fakesmtpd).port + 1).body
    ).fetch('_embedded').fetch('messages').map { |h| h['_links']['self']['href'] }
  end

  def clear_mail_messages
    delete_request(path: '/messages', port: Mtbb.server(:fakesmtpd).port + 1)
  end

  def last_message
    message = JSON.parse(get_request(
      path: current_mail_messages.last, port: Mtbb.server(:fakesmtpd).port + 1
    ).body)
    return Mail.new if message['body'].nil?
    return Mail.new(message['body'].join("\n"))
  end

  def last_message_header(header_name)
    last_message[header_name].to_s
  end
end
