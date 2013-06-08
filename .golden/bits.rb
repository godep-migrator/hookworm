require 'net/http'

GOLD = "\033\[33;1m"
RESET = "\033\[0m"
GREEN = "\033\[32m"
RED = "\033\[31m"
BRIGHT_GREEN = "\033\[32;1m"
BRIGHT_RED = "\033\[31;1m"

def announce!(something)
  $stderr.puts "#{GOLD}golden#{RESET}: #{GREEN}#{something}#{RESET}"
end

class MiniTestReporter
  def puts(*args)
    args.each { |arg| announce! arg }
  end

  alias print puts
end

def post_request(options = {})
  port = options.fetch(:port)
  request = Net::HTTP::Post.new('/')
  request.content_type = 'application/x-www-form-urlencoded'
  request.body = options.fetch(:body)

  response = Net::HTTP.start('localhost', port) do |http|
    http.request(request)
  end

  Integer(response.code)
end

def payload(name)
  filename = {
    plain: 'payload.json',
  }.fetch(name)
  "payload=#{
    File.read(File.expand_path("../../sampledata/#{filename}", __FILE__))
  }"
end
