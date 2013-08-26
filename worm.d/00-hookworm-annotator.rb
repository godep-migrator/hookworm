#!/usr/bin/env ruby

require 'json'
require_relative '.hookworm_base'

class HookwormAnnotator
  include HookwormBase

  private

  def handle_github
    payload = JSON.parse(input_stream.read, symbolize_names: true)
    annotate_github_payload!(payload)
    output_stream.puts JSON.pretty_generate(payload)
  end

  def handle_travis
    output_stream.puts input_stream.read
  end

  def annotate_github_payload!(github_payload)
    HookwormGithubPayloadAnnotator.new(cfg).annotate(github_payload)
  end
end

class HookwormGithubPayloadAnnotator
  PULL_REQUEST_MESSAGE_RE = /Merge pull request #[0-9]+ from.*/

  def initialize(cfg)
    @cfg = cfg
  end

  def annotate(payload)
    annotate_is_pr_merge(payload)
    payload
  end

  private

  def annotate_is_pr_merge(payload)
    payload[:is_pr_merge] = (
      (payload[:commits] || []).length > 1 &&
      !!PULL_REQUEST_MESSAGE_RE.match(payload[:head_commit][:message])
    )
  end
end

if $0 == __FILE__
  HookwormAnnotator.new.run!(ARGV)
end
