require 'open3'
require 'tmpdir'

require_relative 'test_helper'

describe 'hookworm logger' do
  include HookwormJunkDrawer

  def handle(stdin_string, args)
    ENV['HOOKWORM_WORKING_DIR'] = @tempdir
    command = [
      'ruby',
      File.expand_path('../../worm.d/00-hookworm-logger.rb', __FILE__)
    ] + args
    out, err = '', ''
    exit_status = 1

    Open3.popen3(*command) do |stdin, stdout, stderr, wait_thr|
      stdin.write stdin_string
      stdin.close
      out << stdout.read
      err << stderr.read
      exit_status = wait_thr.value
    end

    [exit_status == 0, out, err]
  end

  def handler_config(fizz, working_dir)
    {
      'fizz' => fizz,
      'working_dir' => working_dir
    }
  end

  before do
    @fizz = rand(0..999)
    @tempdir = Dir.mktmpdir
    @handler_config = handler_config(@fizz, @tempdir)
  end

  after do
    if @tempdir
      FileUtils.rm_rf(@tempdir)
    end
  end

  describe 'when given an invalid command' do
    it 'explodes' do
      Dir.chdir(@tempdir) do
        handle('', %w(fribble)).first.must_equal false
      end
    end
  end

  describe 'when configuring' do
    it 'writes JSON from stdin to a config file' do
      Dir.chdir(@tempdir) do
        handle(JSON.dump(@handler_config), %w(configure))
      end
      File.exists?("#{@tempdir}/00-hookworm-logger.rb.cfg.json").must_equal true
    end
  end

  describe 'when handling github payloads' do
    before do
      @github_payload = github_payload_hash('pull_request')
      @github_payload[:repository].merge!({id: @fizz})
      Dir.chdir(@tempdir) do
        handle(JSON.dump(@handler_config), %w(configure))
      end
    end

    it 'logs if the payload is a pull request merge' do
      Dir.chdir(@tempdir) do
        err = handle(JSON.dump(@github_payload), %w(handle github)).last
        err.must_match(/Pull request merge\? true/)
      end
    end

    it 'echoes the payload unaltered' do
      Dir.chdir(@tempdir) do
        out = handle(JSON.dump(@github_payload), %w(handle github))[1]
        JSON.parse(out, symbolize_names: true).must_equal @github_payload
      end
    end
  end

  describe 'when handling travis payloads' do
    before do
      @travis_payload = travis_payload_hash('success')
      Dir.chdir(@tempdir) do
        handle(JSON.dump(@handler_config), %w(configure))
      end
    end

    it 'echoes the payload unaltered' do
      Dir.chdir(@tempdir) do
        out = handle(JSON.dump(@travis_payload), %w(handle travis))[1]
        JSON.parse(out, symbolize_names: true).must_equal @travis_payload
      end
    end
  end
end
