require 'socket'
require 'json'

class FakeMail
  attr_reader :port, :server, :message_dir

  def initialize(options = {})
    @port = options.fetch(:port)
    @message_dir = options.fetch(:dir)
  end

  def start
    @server = TCPServer.new(port)
    puts "FakeMail Serving on #{port}, " <<
         "writing messages to #{message_dir.inspect}"
    loop do
      begin
        serve(server.accept)
      rescue => e
        puts "WAT: #{e.class.name} #{e.message}"
      end
    end
  end

  def serve(client)
    class << client
      def getline
        line = gets
        line.chomp! unless line.nil?
        line
      end
    end
    client.puts '220 localhost FakeMail ready ESMTP'
    helo = client.getline
    puts "Helo: #{helo.inspect}"

    if helo =~ /^EHLO\s+/
      puts 'Seen an EHLO'
      client.puts '250-localhost only has this one extension'
      client.puts '250 HELP'
    end

    from = client.getline
    client.puts('250 OK')
    puts "From: #{from.inspect}"

    to_list = []
    loop do
      to = client.getline
      break if to.nil?

      if to =~ /^DATA/
        client.puts '354 Lemme have it'
        break
      else
        puts "To: #{to.inspect}"
        to_list << to
        client.puts '250 OK'
      end
    end

    lines = []
    loop do
      line = client.getline
      break if line.nil? || line == '.'
      lines << line
      puts "+ #{line}"
    end

    client.puts '250 OK'
    client.gets
    client.puts '221 Buhbye'
    client.close
    puts 'Another one bytes the dust.'

    record(from, to_list, lines.join("\n"))
  end

  def record(from, to_list, body)
    now = Time.now.utc.strftime('%Y%m%d%H%M%S')
    outfile = File.join(message_dir, "fake-mail-#{now}.json")
    File.open(outfile, 'w') do |f|
      f.write JSON.pretty_generate(
        from: from,
        to_list: to_list,
        body: body,
      )
    end
  end
end

def main(argv = [].freeze)
  FakeMail.new(port: argv.fetch(0), dir: argv.fetch(1)).start
end

if __FILE__ == $0
  main(ARGV)
end
