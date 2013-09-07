#!/usr/bin/env ruby
# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8
#+ #### Hookworm Annotator
#+
#+ The annotator is responsible for adding fields to the incoming payloads so
#+ that subsequent handlers do not have to duplicate decision-making logic.
#+
#+ ##### GitHub payload annotation
#+ GitHub payloads are given the following additional fields dependending on
#+ the presence of certain options.
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
    annotated_payload = annotate_github_payload!(payload)
    output_stream.puts JSON.pretty_generate(annotated_payload)
    0
  end

  def handle_travis
    78
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
    payload.merge(
      is_pr_merge: pr_merge?(payload),
      is_watched_branch: watched_branch?(payload[:ref]),
      has_watched_path: watched_path?(payload)
    )
  end

  private

  def pr_merge?(payload)
    truth = (
      (payload[:commits] || []).length > 1 &&
      !!PULL_REQUEST_MESSAGE_RE.match(payload[:head_commit][:message])
    )
    log.debug { "pull request merge? #{truth}" }
    truth
  end

  def watched_branch?(ref)
    sans_refs_heads = ref.sub(%r{^refs/heads/}, '')
    watched_branches.each do |br|
      if sans_refs_heads =~ br
        log.debug { "#{sans_refs_heads} =~ #{br.inspect} -> true" }
        return true
      end
      log.debug { "#{sans_refs_heads} =~ #{br.inspect} -> false" }
    end
    false
  end

  def watched_branches
    @watched_branches ||= watched_branch_strings.map { |wb| /#{wb}/ }
  end

  def watched_branch_strings
    ((cfg[:worm_flags] || {})[:watched_branches] || '').cleanquotes.commasplit
  end

  def watched_path?(payload)
    watched_paths.each do |wp|
      payload_paths(payload).each do |path|
        if path =~ wp
          log.debug { "#{path} =~ #{wp.inspect} -> true" }
          return true
        end
        log.debug { "#{path} =~ #{wp.inspect} -> false" }
      end
    end
    false
  end

  def watched_paths
    @watched_paths ||= watched_path_strings.map { |wp| /#{wp}/ }
  end

  def watched_path_strings
    ((cfg[:worm_flags] || {})[:watched_paths] || '').cleanquotes.commasplit
  end

  def payload_paths(payload)
    paths = []
    commits = payload[:commits] || []
    commits << payload[:head_commit]

    commits.each_with_index do |commit, i|
      next if payload[:is_pr_merge] && i == 0

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

  def log
    @log ||= Logger.new(log_stream)
  end

  def log_stream
    $stderr.set_encoding('UTF-8')
  end
end

exit HookwormAnnotator.new.run!(ARGV) if $PROGRAM_NAME == __FILE__
