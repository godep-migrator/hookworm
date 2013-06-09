require 'net/http'

YELLOW = "\033\[33m"
RESET = "\033\[0m"
GREEN = "\033\[32m"
RED = "\033\[31m"
MAGENTA = "\033\[35m"
BRIGHT_GREEN = "\033\[32;1m"
BRIGHT_RED = "\033\[31;1m"
BRIGHT_MAGENTA = "\033\[35;1m"
BRIGHT_YELLOW = "\033\[33;1m"

class MiniTestReporter
  def puts(*args)
    args.each { |arg| announce! arg }
  end

  alias print puts
end

module Bits
  def announce!(something)
    $stderr.puts "#{MAGENTA}runtests#{RESET}: " <<
                 "#{YELLOW}#{something}#{RESET}"
  end

  def post_request(options = {})
    port = options.fetch(:port)
    request = Net::HTTP::Post.new('/')
    request.content_type = 'application/x-www-form-urlencoded'
    request.body = options.fetch(:body)

    Net::HTTP.start('localhost', port) do |http|
      http.request(request)
    end
  end

  def payload(name)
    "payload=#{URI.escape(File.read(payload_file(name)))}"
  end

  def payload_file(name)
    File.expand_path("../../sampledata/#{name.to_s}.json", __FILE__)
  end
end
