require_relative 'test_helper'

require 'open3'
require 'tmpdir'

describe 'hookworm logger' do
  include Annunciation
  include NetThings
  include Open3

  def handle(stdin_string, args)
    ENV['HOOKWORM_WORKING_DIR'] = @tempdir
    command = [
      'go', 'run',
      File.expand_path('../../worm.d/10-hookworm-rogue-commit-handler.go', __FILE__)
    ] + args
    out_err = ''
    exit_status = 1

    popen2e(*command) do |stdin, stdout_stderr, wait_thr|
      stdin.write stdin_string
      stdin.close
      out_err << stdout_stderr.read
      exit_status = wait_thr.value
    end

    [exit_status == 0, out_err]
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
      out_err = ''
      Dir.chdir(@tempdir) do
        out_err = handle(JSON.dump(@handler_config), %w(configure)).last
      end
      File.exists?("#{@tempdir}/10-hookworm-rogue-commit-handler.go.cfg.json").must_equal true
    end
  end
end
