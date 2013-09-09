#!/usr/bin/env ruby
# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8
#+ #### Hookworm Build Index Handler
#+

require 'json'
require 'fileutils'
require 'time'
require_relative '.hookworm_base'

class HookwormBuildIndexHandler
  include HookwormBase

  private

  def handle_travis
    payload = JSON.parse(input_stream.read, symbolize_names: true)
    HookwormTravisPayloadBuildIndexer.new(cfg, log).index(payload)
    0
  rescue HookwormTravisPayloadBuildIndexer::InvalidPayload => e
    log.error { e }
  ensure
    output_stream.puts JSON.pretty_generate(payload)
    0
  end
end

class HookwormTravisPayloadBuildIndexer
  InvalidPayload = Class.new(StandardError)

  attr_reader :cfg, :log

  include FileUtils

  def initialize(cfg, log)
    @cfg = cfg
    @log = log
  end

  def index(payload)
    mkdir_p(index_prefix)
    repo_slug = (payload[:repository] || {})[:slug]
    unless repo_slug
      raise InvalidPayload.new("Invalid repo slug #{repo_slug.inspect}")
    end
    build_id_path = payload_build_id_path(repo_slug, payload)
    mkdir_p(File.dirname(build_id_path))
    write_payload(payload, build_id_path)
    link_all_category_paths(payload, repo_slug, build_id_path)
  end

  private

  def write_payload(payload, path)
    File.open(path, 'w:UTF-8') do |f|
      f.puts JSON.pretty_generate(payload)
    end
  end

  def payload_build_id_path(repo_slug, payload)
    "#{index_prefix}/#{repo_slug}/builds/#{payload[:id]}.json"
  end

  def link_all_category_paths(payload, repo_slug, build_id_path)
    link_by_number_path(payload, repo_slug, build_id_path)
    link_by_commit_paths(payload, repo_slug, build_id_path)
    link_by_datetime_path(payload, repo_slug, build_id_path)
    link_latest_build(repo_slug)
  end

  def link_by_number_path(payload, repo_slug, build_id_path)
    mkdir_p_ln_sf(
      build_id_path,
      "#{index_prefix}/#{repo_slug}/builds/_by_number/#{payload[:number]}.json"
    )
  end

  def link_by_commit_paths(payload, repo_slug, build_id_path)
    mkdir_p_ln_sf(
      build_id_path,
      "#{index_prefix}/#{repo_slug}/builds/_by_commit/#{payload[:commit]}.json"
    )
    mkdir_p_ln_sf(
      build_id_path,
      "#{index_prefix}/#{repo_slug}/builds/_by_commit/" <<
        "#{payload[:commit][0, 7]}.json"
    )
  end

  def link_by_datetime_path(payload, repo_slug, build_id_path)
    timestamp = Time.parse(payload[:finished_at]).strftime('%Y%m%d_%H%M%S')
    mkdir_p_ln_sf(
      build_id_path,
      "#{index_prefix}/#{repo_slug}/builds/_by_datetime/#{timestamp}.json"
    )
  end

  def link_latest_build(repo_slug)
    builds_glob = "#{index_prefix}/#{repo_slug}/builds/_by_datetime/*.json"
    mkdir_p_ln_sf(
      Dir.glob(builds_glob).sort.last,
      "#{index_prefix}/#{repo_slug}/builds/_latest.json"
    )
  end

  def mkdir_p_ln_sf(src, dest)
    mkdir_p(File.dirname(dest))
    log.debug { "Linking #{src} -> #{dest}" }
    ln_sf(src, dest)
  end

  def index_prefix
    @index_prefix ||= "#{@cfg[:static_dir]}/build-index"
  end
end

exit HookwormBuildIndexHandler.new.run!(ARGV) if $PROGRAM_NAME == __FILE__
