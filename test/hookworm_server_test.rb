require_relative 'test_helper'

$servers.each do |name,server|
  describe "#{name} server receiving hook payloads" do
    include NetThings

    it 'accepts POSTs' do
      post_github_payload(server.port, :valid).first.code.must_equal '204'
    end
  end
end

describe 'when receiving a payload for a watched branch' do
  include NetThings

  before do
    @sent_messages = post_github_payload($servers[:debug].port, :rogue).last
  end

  it 'sends a rogue commit email' do
    @sent_messages.wont_be_empty
  end
end

describe 'when receiving a payload for an unwatched branch' do
  include NetThings

  before do
    @sent_messages = post_github_payload(
      $servers[:debug].port, :rogue_unwatched_branch
    ).last
  end

  it 'does not send a rogue commit email' do
    @sent_messages.must_be_empty
  end
end

describe 'when receiving a payload for an unwatched path' do
  include NetThings

  before do
    @sent_messages = post_github_payload(
      $servers[:debug].port, :rogue_unwatched_path
    ).last
  end

  it 'does not send a rogue commit email' do
    @sent_messages.must_be_empty
  end
end

describe 'rogue commit emails' do
  include NetThings

  before do
    @rogue_response ||= post_github_payload($servers[:debug].port, :rogue).first
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

describe 'when receiving a payload for a pull request merges' do
  describe 'without signoff' do
    it 'sends a rogue pull request email'
  end
end
