#!/usr/bin/env ruby
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
#+ expressions, e.g.: `watched_branches='^master$,^release_[0-9]'`.  If a commit
#+ payload is received that was not the result of a pull request merge and the
#+ Hookworm Annotator handler has determined that the branch name matches any
#+ of the entries in `watched_branches`, then a rogue commit email will be sent.
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
#+ `email_recipients='devs@company.example.com,project-distro+hookworm@partner.example.net'`
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
    return 0
  end

  def handle_travis
    return 78
  end
end

class Judger
  attr_reader :cfg

  def initialize(cfg)
    @cfg = cfg
  end

  def judge_payload(payload)
    if !payload[:is_watched_branch]
      if debug?
        log.info("#{payload[:ref]} is not a watched branch, yay!")
      end

      return
    end

    if debug?
      log.info("#{payload[:ref]} is a watched branch!")
    end

    hcid = payload[:head_commit][:id]
    if payload[:is_pr_merge]
      if debug?
        log.info("#{hcid} is a pull request merge, yay!")
      end

      return
    end

    if debug?
      log.info("#{hcid} is not a pull request merge!")
    end

    if !payload[:has_watched_path]
      if debug?
        log.info("#{hcid} does not contain watched paths, yay!")
      end

      return
    end

    if debug?
      log.info("#{hcid} contains watched paths!")
    end

    begin
      send_rogue_commit_email!(payload)
      if debug?
        log.info("Sent rogue commit email to #{recipients}")
      end
    rescue => e
      log.error("#{e.class.name} #{e.message}")
      if debug?
        log.error(e.backtrace.join("\n"))
      end
      raise e
    end
  end

  def send_rogue_commit_email!(payload)
    log.warn("WARNING rogue commit! #{payload}, head commit: #{payload[:head_commit]}")

    if recipients.empty?
      log.warn("No email recipients specified, so no emailing!")
      return
    end

    email_body = render_email(payload)
    if debug?
      log.info("Email message:\n#{email_body}\n")
    end

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
    @watched_branches ||= (cfg[:worm_flags][:watched_branches] || '').commasplit
  end

  def watched_paths
    @watched_paths ||= (cfg[:worm_flags][:watched_paths] || '').commasplit
  end

  def emailer
    @emailer ||= Emailer.new(cfg[:worm_flags][:email_uri])
  end

  def debug?
    @debug ||= cfg[:debug]
  end

  def log
    @log ||= Logger.new($hookworm_stderr || $stderr)
  end

  class EmailRenderContext
    def initialize(judger, payload, tmpl)
      @cfg = judger.cfg
      @payload = payload
      @tmpl = tmpl
      @from = judger.fromaddr
      @recipients = judger.recipients.join(', ')
      @date = Time.now.utc.rfc2822
      @message_id = Time.now.strftime('%s%9N')
      @hostname = Socket.gethostname
      @repo = "#{payload[:repository][:owner][:name]}/#{payload[:repository][:name]}"
      @ref = payload[:ref]
      @watched_branches = judger.watched_branches
      @watched_paths = judger.watched_paths
      @repo_url = payload[:repository][:url]
      hc = payload[:head_commit]
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
    Net::SMTP.start(@email_uri.host, @email_uri.port,
                    @email_uri.user, @email_uri.password,
                    @email_uri.user ? :plain : nil) do |smtp|
      if @email_uri.scheme == 'smtps'
        smtp.enable_ssl
      end
      smtp.send_message(msg, from, to)
    end
  end
end

class String
  def commasplit
    self.split(',').map(&:strip)
  end

  def to_plaintext
    self.gsub(/\n/, "\\n").gsub(/\t/, "\\t")
  end

  def to_html
    self.gsub(/\n/, "<br />").gsub(/\t/, "    ")
  end
end

if $0 == __FILE__
  exit HookwormRogueCommitHandler.new.run!(ARGV)
end
