require_relative 'test_helper'

describe 'hookworm logger' do
  include HookwormJunkDrawer

  def base_command
    [
      'ruby',
      File.expand_path('../../worm.d/10-hookworm-logger.rb', __FILE__)
    ]
  end

  def handler_config(fizz, working_dir)
    {
      'fizz' => fizz,
      'working_dir' => working_dir
    }
  end

  before do
    @fizz = rand(0..999)
    @handler_config = handler_config(@fizz, @tempdir)
  end

  after { @tempdir = nil }

  describe 'when given an invalid command' do
    it 'explodes' do
      handle('', %w(fribble)).first.exitstatus.wont_equal 0
    end
  end

  describe 'when configuring' do
    it 'writes JSON from stdin to a config file' do
      handle(JSON.dump(@handler_config), %w(configure))
      File.exists?("#{@tempdir}/10-hookworm-logger.rb.cfg.json").must_equal true
    end
  end

  describe 'when handling github payloads' do
    before do
      @github_payload = github_payload_hash('pull_request')
      @github_payload[:repository].merge!({id: @fizz})
      handle(JSON.dump(@handler_config), %w(configure))
    end

    it 'logs if the payload is a pull request merge' do
      err = handle(JSON.dump(@github_payload), %w(handle github)).last
      err.must_match(/Pull request merge\? true/)
    end

    it 'echoes the payload unaltered' do
      out = handle(JSON.dump(@github_payload), %w(handle github))[1]
      JSON.parse(out, symbolize_names: true).must_equal @github_payload
    end
  end

  describe 'when handling travis payloads' do
    before do
      @travis_payload = travis_payload_hash('success')
      handle(JSON.dump(@handler_config), %w(configure))
    end

    it 'echoes the payload unaltered' do
      out = handle(JSON.dump(@travis_payload), %w(handle travis))[1]
      JSON.parse(out, symbolize_names: true).must_equal @travis_payload
    end
  end
end
