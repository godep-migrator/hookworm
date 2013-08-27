#!/usr/bin/env ruby
#+ #### Hookworm Annotator
#+
#+ The annotator is responsible for adding fields to the incoming payloads so
#+ that subsequent handlers do not have to duplicate decision-making logic.
#+
#+ ##### GitHub payload annotation
#+ GitHub payloads are given the following additional fields dependending on the
#+ presence of certain options.
#+
#+ ###### `is_pr_merge`
#+ Is the payload the result of a pull request merge?
#+
#+ ###### `is_watched_branch`
#+ Is the payload for a branch that is "watched", depending on the presence of
#+ the `watched_branches` postfix keyword argument?
#+
#+ ###### `has_watched_path`
#+ Does the payload contain changes to a "watched" path, depending on the
#+ presence of the `watched_paths` postfix keyword argument?

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
    payload[:has_watched_path] = watched_path?(payload)
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

  def watched_path?(payload)
    watched_paths.each do |wp|
      payload_paths(payload).each do |path|
        if path =~ wp
          return true
        end
      end
    end
    false
  end

  def watched_paths
    @watched_paths ||= watched_path_strings.map { |wp| %r{#{wp}} }
  end

  def watched_path_strings
    ((cfg[:worm_flags] || {})[:watched_paths] || '').split(',')
  end

  def payload_paths(payload)
    paths = []
    commits = payload[:commits] || []
    commits << payload[:head_commit]

    commits.each_with_index do |commit, i|
      if payload[:is_pr_merge] && i == 0
        next
      end

      paths += commit_paths(commit)
    end

    paths
  end

  def commit_paths(commit)
    path_set = {}

    [commit[:added], commit[:removed], commit[:modified]].each do |path_list|
      path_list.each do |path|
        path_set[path] = true
      end
    end

    path_set.keys.sort
  end
end

if $0 == __FILE__
  HookwormAnnotator.new.run!(ARGV)
end
