# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8

require_relative 'test_helper'

describe 'hookworm build index handler' do
  include HookwormJunkDrawer

  def base_command
    [
      'ruby',
      File.expand_path(
        '../../worm.d/30-hookworm-build-index-handler.rb', __FILE__
      )
    ]
  end

  before do
    @fizz = rand(0..999)
    @handler_config = handler_config(@fizz, tempdir)
    @static_dir = @handler_config['static_dir']
  end

  after { @tempdir = nil }

  describe 'when given an invalid command' do
    it 'explodes' do
      handle('', %w(fribble)).first.exitstatus.wont_equal 0
    end
  end

  describe 'when configuring' do
    it 'writes JSON from stdin to a config file' do
      handle(JSON.dump(@handler_config), %w(configure)).last
      File.exists?("#{@tempdir}/30-hookworm-build-index-handler.rb.cfg.json")
        .must_equal true
    end
  end

  describe 'when handling github payloads' do
    before do
      @github_payload = github_payload_hash('pull_request')
      handle(JSON.dump(@handler_config), %w(configure))
    end

    it 'exits 78' do
      ps = handle(JSON.dump(@github_payload), %w(handle github)).first
      ps.exitstatus.must_equal 78
    end
  end

  describe 'when handling travis payloads' do
    def build_index_prefix
      "#{@static_dir}/build-index/meatballhat/fuzzy-octo-archer"
    end

    def build_id_path
      "#{build_index_prefix}/builds/10180613.json"
    end

    def build_number_path
      "#{build_index_prefix}/builds/_by_number/7.json"
    end

    def latest_build_path
      "#{build_index_prefix}/builds/_latest.json"
    end

    def commit_path
      "#{build_index_prefix}/builds/_by_commit/" <<
        'ad281ff8a9d7be26da84a65147b42ca7f6cf6857.json'
    end

    def short_commit_path
      "#{build_index_prefix}/builds/_by_commit/ad281ff.json"
    end

    def datetime_path
      "#{build_index_prefix}/builds/_by_datetime/20130814_022637.json"
    end

    before do
      @travis_payload = travis_payload_hash('annotated_success')
      handle(JSON.dump(@handler_config), %w(configure))
    end

    %w(
      build_id
      build_number
      commit
      datetime
      latest_build
      short_commit
    ).each do |path_type|
      it "stores the payload to the #{path_type} path" do
        handle(JSON.dump(@travis_payload), %w(handle travis))
        path = send(:"#{path_type}_path")
        File.exists?(path).must_equal true
        JSON.parse(File.read(path))['id'].must_equal 10_180_613
      end
    end
  end
end
