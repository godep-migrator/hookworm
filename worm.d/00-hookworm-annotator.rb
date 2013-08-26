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
  attr_reader :cfg

  PULL_REQUEST_MESSAGE_RE = /Merge pull request #[0-9]+ from.*/

  def initialize(cfg)
    @cfg = cfg
  end

  def annotate(payload)
    payload[:is_pr_merge] = pr_merge?(payload)
    payload[:is_watched_branch] = watched_branch?(payload[:ref])
    payload
  end

  private

  def pr_merge?(payload)
    (payload[:commits] || []).length > 1 &&
      !!PULL_REQUEST_MESSAGE_RE.match(payload[:head_commit][:message])
  end

  def watched_branch?(ref)
    sans_refs_heads = ref.sub(%r{^refs/heads/}, '')
    watched_branches.each do |br|
      if sans_refs_heads =~ br
        return true
      end
    end
    false
  end

  def watched_branches
    @watched_branches ||= watched_branch_strings.map { |wb| %r{#{wb}} }
  end

  def watched_branch_strings
    ((cfg[:worm_flags] || {})[:watched_branches] || '').split(',')
  end
end

if $0 == __FILE__
  HookwormAnnotator.new.run!(ARGV)
end
