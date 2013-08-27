require 'open3'
require 'tmpdir'

require_relative 'test_helper'

describe 'hookworm logger' do
  include HookwormJunkDrawer

  def base_command
    ["#{@tempdir}/20-hookworm-rogue-commit-handler"]
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
    system "go build -o #{@tempdir}/20-hookworm-rogue-commit-handler " <<
           "#{File.expand_path('../../worm.d/20-hookworm-rogue-commit-handler.go', __FILE__)}"
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
        handle('', %w(fribble)).first.exitstatus.wont_equal 0
      end
    end
  end

  describe 'when configuring' do
    it 'writes JSON from stdin to a config file' do
      Dir.chdir(@tempdir) do
        handle(JSON.dump(@handler_config), %w(configure)).last
      end
      File.exists?("#{@tempdir}/20-hookworm-rogue-commit-handler.go.cfg.json").must_equal true
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

    it 'exits 78' do
      Dir.chdir(@tempdir) do
        ps = handle(JSON.dump(@travis_payload), %w(handle travis)).first
        ps.exitstatus.must_equal 78
      end
    end
  end
end
