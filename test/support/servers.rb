require 'fileutils'
require 'time'
require 'fakesmtpd'

require_relative 'hookworm_server_runner'

$fakesmtpd_server = FakeSMTPd::Runner.new(
  port: "#{rand(13100..13109)}",
  dir: File.expand_path('../../../log/emails', __FILE__),
  pidfile: File.expand_path('../../../log/fakesmtpd.pid', __FILE__),
  logfile: File.expand_path('../../../log/fakesmtpd.log', __FILE__)
)

$servers = {
  null: HookwormServerRunner.new(
    '-a' => ":#{rand(12100..12109)}",
    '-P' => File.expand_path('../../../log/hookworm-server-null.pid', __FILE__),
    '-D' => File.expand_path(
      "../../../log/hookworm-null-#{Time.now.utc.to_i}-#{$$}", __FILE__
    ),
    start: Time.now.utc,
  ),
  debug: HookwormServerRunner.new(
    '-a' => ":#{rand(12110..12119)}",
    '-d' => nil,
    '-P' => File.expand_path('../../../log/hookworm-server-debug.pid', __FILE__),
    '-b' => '^master$,^develop$',
    '-p' => '\.go$,\.json$',
    '-e' => "smtp://localhost:#{$fakesmtpd_server.port}",
    '-f' => 'hookworm-runtests@testing.local',
    '-r' => 'hookworm-self@testing.local',
    '-D' => File.expand_path(
      "../../../log/hookworm-debug-#{Time.now.utc.to_i}-#{$$}", __FILE__
    ),
    '-W' => File.expand_path('../../../worm.d', __FILE__),
    start: Time.now.utc,
  )
}

def stop_servers
  $fakesmtpd_server.stop
  $servers.each do |_,server|
    server.stop
  end
end

def start_servers
  Dir.chdir(File.expand_path('../../../', __FILE__)) do
    FileUtils.mkdir_p('./log')
    $fakesmtpd_server.start
    $servers.each do |_,runner|
      runner.start
    end
  end
end
