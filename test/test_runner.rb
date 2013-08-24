# vim:fileencoding=utf-8

def main(argv = [].freeze)
  ENV['IN_TEST_RUNNER'] = '1'
  require_relative 'test_helper'

  include Annunciation

  Dir.glob("#{File.expand_path('../', __FILE__)}/**/*_test.rb").each do |f|
    require f
  end

  start_servers
  announce! 'Started servers'
  at_exit do
    stop_servers
    announce! 'Stopped servers'
  end
  exit_code = 1

  Dir.chdir(File.expand_path('../../log', __FILE__)) do
    MiniTest::Unit.output = MiniTestReporter.new
    exit_code = MiniTest::Unit.new.run(argv) || 1
  end

  if exit_code == 0
    $stderr.puts BRIGHT_GREEN
    $stderr.puts <<-EOF.gsub(/^ {4}/, '')
      ✓✓      ✓✓ ✓✓✓✓ ✓✓    ✓✓
      ✓✓  ✓✓  ✓✓  ✓✓  ✓✓✓   ✓✓
      ✓✓  ✓✓  ✓✓  ✓✓  ✓✓✓✓  ✓✓
      ✓✓  ✓✓  ✓✓  ✓✓  ✓✓ ✓✓ ✓✓
      ✓✓  ✓✓  ✓✓  ✓✓  ✓✓  ✓✓✓✓
      ✓✓  ✓✓  ✓✓  ✓✓  ✓✓   ✓✓✓
       ✓✓✓  ✓✓✓  ✓✓✓✓ ✓✓    ✓✓
    EOF
    $stderr.puts RESET
  else
    $stderr.puts BRIGHT_RED
    $stderr.puts <<-EOF.gsub(/^ {4}/, '')
      ✘✘✘✘✘✘✘✘    ✘✘✘    ✘✘✘✘ ✘✘
      ✘✘         ✘✘ ✘✘    ✘✘  ✘✘
      ✘✘        ✘✘   ✘✘   ✘✘  ✘✘
      ✘✘✘✘✘✘   ✘✘     ✘✘  ✘✘  ✘✘
      ✘✘       ✘✘✘✘✘✘✘✘✘  ✘✘  ✘✘
      ✘✘       ✘✘     ✘✘  ✘✘  ✘✘
      ✘✘       ✘✘     ✘✘ ✘✘✘✘ ✘✘✘✘✘✘✘✘
    EOF
    $stderr.puts RESET

    $servers.each { |_,server| server.dump_log }
  end

  if ENV['HOLD_ON_A_SEC']
    print 'Holding on a sec... [Enter] '
    gets
  end
  exit exit_code
end

if __FILE__ == $0
  main(ARGV)
end
