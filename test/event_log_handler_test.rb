$LOAD_PATH.unshift(File.expand_path('../../', __FILE__))

require 'minitest/autorun'
require 'stringio'
require 'tmpdir'
require '.golden/bits'
require 'worm.d/00-event-log-handler'

describe EventLogHandler do
  include Bits

  def handler
    EventLogHandler.new
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
      proc { handler.run!('fribble') }.must_raise NoMethodError
    end
  end

  describe 'when configuring' do
    before do
      @handler = handler
      $stdin = StringIO.new(JSON.dump(@handler_config))
    end

    it 'writes JSON from stdin to a config file' do
      @handler.run!('configure')
      JSON.parse(File.read(@handler.send(:cfg_file)))['fizz'].must_equal @fizz
    end
  end

  describe 'when handling github payloads' do
    before do
      @old_stderr = $stderr
      @handler = handler
      @github_payload = payload_hash('pull_request')
      @github_payload[:repository].merge!({id: @fizz})
      $stdin = StringIO.new(JSON.dump(@handler_config))
      @handler.run!('configure')
      $stdin = StringIO.new(JSON.dump(@github_payload))
      $stderr = StringIO.new
      @log = Logger.new($stderr)
      @handler.instance_variable_set(:@log, @log)
    end

    after do
      if @old_stderr
        $stderr = @old_stderr
      end
    end

    it 'logs if the payload is a pull request merge' do
      @handler.run!('handle', 'github')
      $stderr.seek(0)
      $stderr.read.must_match(/Pull request merge\? true/)
    end
  end
end
