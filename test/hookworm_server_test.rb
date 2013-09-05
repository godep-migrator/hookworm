require_relative 'test_helper'

Mtbb::SERVERS.each do |name, server|
  next if name == :fakesmtpd

  describe "#{name} server receiving hook payloads" do
    include HookwormJunkDrawer

    it 'accepts POSTs' do
      post_github_payload(server.port, :valid).first.code.must_equal '204'
    end
  end

  describe "#{name} ancillary pages" do
    include HookwormJunkDrawer

    it 'has a test page' do
      get_request(port: server.port, path: '/debug/test').code.must_equal '200'
    end

    it 'has a blank page' do
      get_request(port: server.port, path: '/blank').code.must_equal '204'
    end

    it 'has an index page' do
      get_request(port: server.port, path: '/').code.must_equal '200'
    end

    it 'has a favicon' do
      get_request(port: server.port, path: '/favicon.ico').code.must_equal '200'
    end
  end
end

describe 'when receiving a payload for a watched branch' do
  include HookwormJunkDrawer

  before do
    @sent_messages = post_github_payload(Mtbb.server(:debug).port, :rogue).last
  end

  it 'sends a rogue commit email' do
    @sent_messages.wont_be_empty
  end
end

describe 'when receiving a payload for an unwatched branch' do
  include HookwormJunkDrawer

  before do
    @sent_messages = post_github_payload(
      Mtbb.server(:debug).port, :rogue_unwatched_branch
    ).last
  end

  it 'does not send a rogue commit email' do
    @sent_messages.must_be_empty
  end
end

describe 'when receiving a payload for an unwatched path' do
  include HookwormJunkDrawer

  before do
    @sent_messages = post_github_payload(
      Mtbb.server(:debug).port, :rogue_unwatched_path
    ).last
  end

  it 'does not send a rogue commit email' do
    @sent_messages.must_be_empty
  end
end

describe 'rogue commit emails' do
  include HookwormJunkDrawer

  before do
    @rogue_response ||= post_github_payload(Mtbb.server(:debug).port, :rogue).first
  end

  it 'are multipart' do
    last_message.multipart?.must_equal true
  end

  it 'are sent to the specified recipients' do
    last_message_header('To').must_equal 'hookworm-self@testing.local'
  end

  it 'are sent from the specified sender' do
    last_message_header('From').must_equal 'hookworm-runtests@testing.local'
  end

  it 'have a subject starting with [hookworm]' do
    last_message_header('Subject').must_match(/^\s*\[hookworm\]/)
  end

  it 'have a subject with the commit author name' do
    last_message_header('Subject').must_match(/Dan Buch/)
  end
end
