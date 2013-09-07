#!/usr/bin/env ruby
# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8
#+ #### Hookworm Rogue Commit Handler
#+
#+ The rogue commit handler is specific to GitHub payloads.  It will inspect
#+ a payload in the context of the given `watched_branches` and `watched_paths`
#+ and send a "rogue commit email" to the email recipients given in
#+ `email_recipients` to provide visibility roughly equivalent to those commits
#+ that result from pull request merges.
#+
#+ Because the rogue commit handler is affected by so many arguments, here they
#+ are again with more details about their associated behavior:
#+
#+ ##### `watched_branches`
#+ The `watched_branches` argument should be a comma-delimited list of regular
#+ expressions, e.g.: `watched_branches='^master$,^release_[0-9]'`.  If a
#+ commit payload is received that was not the result of a pull request merge
#+ and the Hookworm Annotator handler has determined that the branch name
#+ matches any of the entries in `watched_branches`, then a rogue commit email
#+ will be sent.
#+
#+ ##### `watched_paths`
#+ The `watched_paths` argument should be a comma-delimited list of regular
#+ expressions, e.g.: `watched_paths='.*\.(go|rb|py)$,bin/.*'`.  If a commit
#+ payload is received that was not the result of a pull request merge and the
#+ Hookworm Annotator handler has determined that one of the commits in the
#+ payload contains a path matching any of the entries in `watched_paths`, then
#+ a rogue commit email will be sent.
#+
#+ ##### `email_from_addr`
#+ The `email_from_addr` is the email address used as the `From` header and
#+ SMTP MAIL address when sending rogue commit emails, e.g.:
#+ `email_from_addr='hookworm-noreply@company.example.com'`.
#+
#+ ##### `email_recipients`
#+ The `email_recipients` argument should be a comma-delimited list of email
#+ addresses (without display name) used in the `To` header and SMTP RCPT
#+ addresses when sending rogue commit emails, e.g.:
#+ `email_recipients='devs@example.com,proj+hookworm@partner.example.net'`
#+
#+ ##### `email_uri`
#+ The `email_uri` argument should be a well-formed URI containing the SMTP
#+ hostname and port and potentially the username and password used for plain
#+ SMTP auth, e.g.:
#+ `email_uri='smtp://hookworm:secret@mailhost.example.com:2025'`

require 'erb'
require 'json'
require 'net/smtp'
require 'time'
require 'uri'
require_relative '.hookworm_base'

class HookwormRogueCommitHandler
  include HookwormBase

  private

  def handle_github
    payload = JSON.parse(input_stream.read, symbolize_names: true)
    begin
      Judger.new(cfg).judge_payload(payload)
    rescue StandardError
      return 1
    ensure
      output_stream.puts JSON.pretty_generate(payload)
    end
    0
  end

  def handle_travis
    78
  end
end

class Judger
  attr_reader :cfg

  def initialize(cfg)
    @cfg = cfg
  end

  def judge_payload(payload)
    return unless is_watched_branch?(payload)
    hcid = payload[:head_commit][:id]
    return if is_pr_merge?(payload, hcid)
    return unless has_watched_path?(payload, hcid)
    safe_send_rogue_commit_email!(payload)
  end

  def new_judge_payload(payload)
    return unless is_watched_branch?(payload)
    hcid = payload[:head_commit][:id]
    return if is_pr_merge?(payload, hcid)
    return unless has_watched_path?(payload, hcid)
    safe_send_rogue_commit_email!(payload)
  end

  def is_watched_branch?(payload)
    unless payload[:is_watched_branch]
      log.debug { "#{payload[:ref]} is not a watched branch, yay!" }
      return false
    end

    log.debug { "#{payload[:ref]} is a watched branch!" }
    true
  end

  def is_pr_merge?(payload, hcid)
    if payload[:is_pr_merge]
      log.info { "#{hcid} is a pull request merge, yay!" }
      return true
    end

    log.debug { "#{hcid} is not a pull request merge!" }
    false
  end

  def has_watched_path?(payload, hcid)
    unless payload[:has_watched_path]
      log.debug { "#{hcid} does not contain watched paths, yay!" }
      return false
    end

    log.debug { "#{hcid} contains watched paths!" }
    true
  end

  def safe_send_rogue_commit_email!(payload)
    send_rogue_commit_email!(payload)
    log.debug { "Sent rogue commit email to #{recipients}" }
  rescue => e
    log.error { "#{e.class.name} #{e.message}" }
    log.debug { e.backtrace.join("\n") }
    raise e
  end

  def send_rogue_commit_email!(payload)
    log.warn do
      "WARNING rogue commit! #{payload}, head commit: #{payload[:head_commit]}"
    end

    if recipients.empty?
      log.warn { 'No email recipients specified, so no emailing!' }
      return
    end

    email_body = render_email(payload)
    log.debug { "Email message:\n#{email_body}" }

    emailer.send(fromaddr, recipients, email_body)
  end

  def render_email(payload)
    EmailRenderContext.new(self, payload, rogue_commit_email_tmpl).render
  end

  def rogue_commit_email_tmpl
    ERB.new(rogue_commit_email_tmpl_string)
  end

  def rogue_commit_email_tmpl_string
    File.read(File.expand_path('../.rogue-commit-email-tmpl.erb', __FILE__))
  end

  def recipients
    @recipients ||= (cfg[:worm_flags][:email_recipients] || '').commasplit
  end

  def fromaddr
    @fromaddr ||= cfg[:worm_flags][:email_from_addr]
  end

  def watched_branches
    @watched_branches ||= (
      cfg[:worm_flags][:watched_branches] || ''
    ).commasplit
  end

  def watched_paths
    @watched_paths ||= (cfg[:worm_flags][:watched_paths] || '').commasplit
  end

  def emailer
    @emailer ||= Emailer.new(cfg[:worm_flags][:email_uri])
  end

  def log
    @log ||= Logger.new($stderr)
  end

  class EmailRenderContext
    def initialize(judger, payload, tmpl)
      assign_judger_vars(judger)
      assign_payload_vars(payload)
      @tmpl = tmpl
      @date = Time.now.utc.rfc2822
      @message_id = Time.now.strftime('%s%9N')
      @hostname = Socket.gethostname
      assign_head_commit_vars(payload[:head_commit])
    end

    def assign_judger_vars(judger)
      @cfg = judger.cfg
      @from = judger.fromaddr
      @recipients = judger.recipients.join(', ')
      @watched_branches = judger.watched_branches
      @watched_paths = judger.watched_paths
    end

    def assign_payload_vars(payload)
      @payload = payload
      @repo = "#{payload[:repository][:owner][:name]}/" <<
              "#{payload[:repository][:name]}"
      @ref = payload[:ref]
      @repo_url = payload[:repository][:url]
    end

    def assign_head_commit_vars(hc)
      @head_commit_id = hc[:id]
      @head_commit_url = hc[:url]
      @head_commit_author = hc[:author][:name]
      @head_commit_committer = hc[:committer][:name]
      @head_commit_message_text = hc[:message].to_plaintext
      @head_commit_message_html = hc[:message].to_html
      @head_commit_timestamp = hc[:timestamp]
    end

    def render
      @tmpl.result(binding)
    end
  end
end

class Emailer
  def initialize(email_uri)
    @email_uri = URI(email_uri)
  end

  def send(from, to, msg)
    Net::SMTP.start(*smtp_args) do |smtp|
      smtp.enable_ssl if @email_uri.scheme == 'smtps'
      smtp.send_message(msg, from, to)
    end
  end

  private

  def smtp_args
    [
      @email_uri.host,
      @email_uri.port,
      @email_uri.user,
      @email_uri.password,
      @email_uri.user ? :plain : nil
    ]
  end
end

exit HookwormRogueCommitHandler.new.run!(ARGV) if $PROGRAM_NAME == __FILE__
