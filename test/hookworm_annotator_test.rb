# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8

require_relative 'test_helper'

describe 'hookworm annotator' do
  include HookwormJunkDrawer

  def base_command
    [
      'ruby',
      File.expand_path('../../worm.d/00-hookworm-annotator.rb', __FILE__)
    ]
  end

  def handler_config(fizz, working_dir)
    {
      'fizz' => fizz,
      'working_dir' => working_dir,
      'worm_flags' => {
        'watched_branches' => '^master$',
        'watched_paths' => '.*',
      }
    }
  end

  before do
    @fizz = rand(0..999)
    @handler_config = handler_config(@fizz, tempdir)
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
      File.exists?("#{@tempdir}/00-hookworm-annotator.rb.cfg.json")
        .must_equal true
    end
  end

  describe 'when handling github payloads' do
    before do
      @github_payload = github_payload_hash('pull_request')
      @github_payload[:repository].merge!({ id: @fizz })
      handle(JSON.dump(@handler_config), %w(configure))
    end

    it 'annotates is_pr_merge' do
      out = handle(JSON.dump(@github_payload), %w(handle github))[1]
      JSON.parse(out, symbolize_names: true)[:is_pr_merge].must_equal true
    end

    it 'annotates is_watched_branch' do
      out = handle(JSON.dump(@github_payload), %w(handle github))[1]
      JSON.parse(out, symbolize_names: true)[:is_watched_branch]
        .must_equal true
    end

    it 'annotates has_watched_path' do
      out = handle(JSON.dump(@github_payload), %w(handle github))[1]
      JSON.parse(out, symbolize_names: true)[:has_watched_path].must_equal true
    end
  end

  describe 'when handling travis payloads' do
    before do
      @travis_payload = travis_payload_hash('success')
      handle(JSON.dump(@handler_config), %w(configure))
    end
  end
end
