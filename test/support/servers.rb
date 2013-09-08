# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8

require 'English'
require 'fileutils'
require 'time'

fakesmtpd_port = rand(13100..13109)
fakesmtpd_message_dir = File.expand_path(
  '../../../.mtbb-artifacts/emails', __FILE__
)
Mtbb.register(
  :fakesmtpd,
  server_name: 'fakesmtpd',
  executable: File.expand_path('../../../fakesmtpd', __FILE__),
  argv: [
    "#{fakesmtpd_port}",
    fakesmtpd_message_dir,
    '-p', File.expand_path('../../../.mtbb-artifacts/fakesmtpd.pid', __FILE__),
    '-l', File.expand_path('../../../.mtbb-artifacts/fakesmtpd.log', __FILE__)
  ],
  port: fakesmtpd_port,
  start: Time.now.utc,
)

null_working_dir = File.expand_path(
  "../../../.mtbb-artifacts/hookworm-null-#{Time.now.utc.to_i}-#{$PID}",
  __FILE__
)
null_port = rand(12100..12109)
Mtbb.register(
  :null,
  server_name: 'hookworm-null',
  executable: "#{ENV['GOPATH'].split(/:/).first}/bin/hookworm-server",
  argv: [
    '-a', ":#{null_port}",
    '-P', File.expand_path(
      '../../../.mtbb-artifacts/hookworm-server-null.pid', __FILE__
    ),
    '-D', null_working_dir,
  ],
  port: null_port,
  start: Time.now.utc,
)

debug_working_dir = File.expand_path(
  "../../../.mtbb-artifacts/hookworm-debug-#{Time.now.utc.to_i}-#{$PID}",
  __FILE__
)
debug_port = rand(12110..12119)
Mtbb.register(
  :debug,
  server_name: 'hookworm-debug',
  executable: "#{ENV['GOPATH'].split(/:/).first}/bin/hookworm-server",
  argv: [
    '-a', ":#{debug_port}",
    '-d',
    '-P', File.expand_path(
      '../../../.mtbb-artifacts/hookworm-server-debug.pid', __FILE__
    ),
    '-D', debug_working_dir,
    '-T', '5',
    '-W', File.expand_path('../../../worm.d', __FILE__),
    'watched_branches=^master$,^develop$',
    'watched_paths=\\.go$,\\.json$',
    "email_uri=smtp://localhost:#{fakesmtpd_port}",
    'email_from_addr=hookworm-runtests@testing.local',
    'email_recipients=hookworm-self@testing.local',
  ],
  port: debug_port,
  start: Time.now.utc,
)

[fakesmtpd_message_dir, null_working_dir, debug_working_dir].each do |dir|
  FileUtils.mkdir_p(dir)
end

at_exit do
  [fakesmtpd_message_dir, null_working_dir, debug_working_dir].each do |dir|
    FileUtils.rm_rf(dir)
  end
end
