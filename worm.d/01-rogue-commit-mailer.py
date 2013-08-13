#!/bin/python/
import smtplib
from os import getlogin
from socket import getfqdn
import sys

watched_branches = [] # can be made into an object or something, should be passed around and not global
payload = {}          # ^^

def main():

    global payload
    payload = Payload()
    message = get_message()

    sender = get_sender(message)
    receivers = get_receivers()

    sendEmail(sender, receivers, message)

'''
serialized the json passed to the command line as a map
'''
def Payload():
    data = sys.argv[1]
    return json.loads(data)

'''
Reaches into 'message' and pulls out the name of the receivers
Or, just gets them from the payload
'''
def get_sender(message):
    payload["Recipients"]


'''
TODO: Ask Dan how 'watched branches is populated'
This should set the global variable watched_branches
'''
def populate_watched_branches():
    pass

'''
TODO: complete
This determines if the branch committed to was a member of the list of watched branched
'''
def is_watched_branch(payload, watchedBranches):

    ref = payload[ref]
    removable = 'ref/heads/'
    position = ref.find(removable)

    if position != -1:
        ref = ref[:position] + ref[position + len(removable):]

    #this needs to be finished off
    for branch in watched_branches:
        pass

    '''
    golang code
    payload.ref.string is passed as ref
        sans = strings.Replace(ref, "ref/heads/", "". 1)
        for branchRe in me.watchedBranches {
            if branchRe.MatchString(sans)
    }
    '''

'''
Wrapper for the smtp protocol
'''
def sendEmail(sender, receivers, message):
    try:
        smtpObj = smtplib.SMTP('localhost')
        smtpObj.sendmail(sender, receivers, message)
        print "Successfully sent email"

    except SMTPException:
        print "Error: unable to send email"

'''
TODO: complete
Called __template_string to get the string that needs templated,
Templates it using the sent json
returns it as the message body
'''
def get_message():
    pass


'''
This is just the raw template
This needs to be actual python-style templating
'''
def __template_string():
    """To: ${Recipients}
    Subject: [hookworm] Rogue commit by ${HeadCommitAuthor} to ${Repo} ${Ref}} (${HeadCommitId})
    Date: ${Date}
    Message-ID: <${MessageId}@${Hostname}>
    List-ID: ${Repo} <hookworm.github.com>
    Content-Type: multipart/alternative;
      boundary="--==ZOMGBOUNDARAAAYYYYY";
      charset=UTF-8
    Content-Transfer-Encoding: 7bit

    ----==ZOMGBOUNDARAAAYYYYY
    Date: ${Date}
    Mime-Version: 1.0
    Content-Type: text/plain; charset=utf8
    Content-Transfer-Encoding: 7bit

    Rogue commit detected!

    Repo      ${Repo}
    Ref       ${Ref}
    Id        ${HeadCommitUrl}
    Author    ${HeadCommitAuthor}
    Committer ${HeadCommitCommitter}
    Timestamp ${HeadCommitTimestamp}
    Message   ${HeadCommitMessageText}

    --
    This email was sent by hookworm:  https://github.com/modcloth-labs/hookworm

    A rogue commit is a commit made directly to a branch that is being watched such
    that only pull requests should be merged into them.

    The configured watched branches are:

    ${WatchedBranches} -

    The configured watched paths are:

    ${WatchedPaths} -


    If you believe this rogue commit email is an error, you should hunt down the
    party responsible for the hookworm instance registered as a WebHook URL in this
    repo's service hook settings (${RepoUrl}/settings/hooks).

    Pretty please submit issues specific to hookworm functionality on github:
    https://github.com/modcloth-labs/hookworm/issues/new

    ----==ZOMGBOUNDARAAAYYYYY
    Date: ${Date}}
    Mime-Version: 1.0
    Content-Type: text/html; charset=utf8
    Content-Transfer-Encoding: 7bit

    <div>
      <h1><a href="${HeadCommitUrl}">Rogue commit detected!</a></h1>

      <table>
        <thead><th></th><th></th></thead>
        <tbody>
          <tr>
            <td style="text-align:right;vertical-align:top;">
              <strong>Repo</strong>:
            </td>
            <td>${Repo}}</td>
          </tr>
          <tr>
            <td style="text-align:right;vertical-align:top;">
              <strong>Ref</strong>:
            </td>
            <td>${Ref}}</td>
          </tr>
          <tr>
            <td style="text-align:right;vertical-align:top;">
              <strong>Id</strong>:
            </td>
            <td><a href="${HeadCommitUrl}">${HeadCommitId}</a></td>
          </tr>
          <tr>
            <td style="text-align:right;vertical-align:top;">
              <strong>Author</strong>:
            </td>
            <td>${HeadCommitAuthor}</td>
          </tr>
          <tr>
            <td style="text-align:right;vertical-align:top;">
              <strong>Committer</strong>:
            </td>
            <td>${HeadCommitCommitter}</td>
          </tr>
          <tr>
            <td style="text-align:right;vertical-align:top;">
              <strong>Timestamp</strong>:
            </td>
            <td>${HeadCommitTimestamp}</td>
          </tr>
          <tr>
            <td style="text-align:right;vertical-align:top;">
              <strong>Message</strong>:
            </td>
            <td>${HeadCommitMessageHtml}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <hr/>
    <div style="font-size:.8em">
      <p>
        This email was sent by
        <a href="https://github.com/modcloth-labs/hookworm">hookworm</a>.
      </p>

      <p>
        A rogue commit is a commit made directly to a branch that is being watched such
        that only pull requests should be merged into them.
      </p>

      <p>
        The configured watched branches are:
      </p>

      <ul>
        ${WatchedBranches}}<li><strong></strong></li>
      </ul>

      <p>
        The configured watched paths are:
      </p>

      <ul>
        ${WatchedPaths}}<li><strong></strong></li>
      </ul>

      <p>
        If you believe this rogue commit email is an error, you should hunt down the
        party responsible for the hookworm instance registered as a WebHook URL in
        this repo's <a href="${RepoUrl}/settings/hooks">service hook settings</a>.
      </p>

      <p>
        Pretty please submit issues specific to hookworm functionality
        <a href="https://github.com/modcloth-labs/hookworm/issues/new">on github</a>
      </p>

    </div>

    ----==ZOMGBOUNDARAAAYYYYY--
    """

if __name__ == '__main__':
    main() 
