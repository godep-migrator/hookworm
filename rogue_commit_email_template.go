package hookworm

import (
	"text/template"
)

var (
	rogueCommitEmailTmpl = template.Must(template.New("email").Parse(`From: {{.From}}
To: {{.Recipients}}
Subject: [hookworm] Rogue commit by {{.HeadCommitAuthor}} to {{.Repo}} {{.Ref}} ({{.HeadCommitId}})
Date: {{.Date}}
Message-ID: <{{.MessageId}}@{{.Hostname}}>
List-ID: {{.Repo}} <hookworm.github.com>
Content-Type: multipart/alternative;
  boundary="--==ZOMGBOUNDARAAAYYYYY";
  charset=UTF-8
Content-Transfer-Encoding: 7bit

----==ZOMGBOUNDARAAAYYYYY
Date: {{.Date}}
Mime-Version: 1.0
Content-Type: text/plain; charset=utf8
Content-Transfer-Encoding: 7bit

Rogue commit detected!

Repo      {{.Repo}}
Ref       {{.Ref}}
Id        {{.HeadCommitUrl}}
Author    {{.HeadCommitAuthor}}
Committer {{.HeadCommitCommitter}}
Timestamp {{.HeadCommitTimestamp}}
Message   {{.HeadCommitMessageText}}

-- 
This email was sent by hookworm:  https://github.com/modcloth-labs/hookworm

A rogue commit is a commit made directly to a branch that has been deemed
sacred or policed such that only pull requests should be merged into them.
The configured policed branches are {{.PolicedBranches}}.

If you believe this rogue commit email is an error, you should hunt down the
party responsible for the hookworm instance registered as a WebHook URL in this
repo's service hook settings ({{.RepoUrl}}/settings/hooks).

Pretty please submit issues specific to hookworm functionality on github:
https://github.com/modcloth-labs/hookworm/issues/new

----==ZOMGBOUNDARAAAYYYYY
Date: {{.Date}}
Mime-Version: 1.0
Content-Type: text/html; charset=utf8
Content-Transfer-Encoding: 7bit

<div>
  <h1><a href="{{.HeadCommitUrl}}">Rogue commit detected!</a></h1>

  <table>
    <thead><th></th><th></th></thead>
    <tbody>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Repo</strong>:
        </td>
        <td>{{.Repo}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Ref</strong>:
        </td>
        <td>{{.Ref}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Id</strong>:
        </td>
        <td><a href="{{.HeadCommitUrl}}">{{.HeadCommitId}}</a></td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Author</strong>:
        </td>
        <td>{{.HeadCommitAuthor}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Committer</strong>:
        </td>
        <td>{{.HeadCommitCommitter}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Timestamp</strong>:
        </td>
        <td>{{.HeadCommitTimestamp}}</td>
      </tr>
      <tr>
        <td style="text-align:right;vertical-align:top;">
          <strong>Message</strong>:
        </td>
        <td>{{.HeadCommitMessageHtml}}</td>
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
    A rogue commit is a commit made directly to a branch that has been deemed
    sacred or policed such that only pull requests should be merged into them.
    The configured policed branches are <strong>{{.PolicedBranches}}</strong>.
  </p>

  <p>
    If you believe this rogue commit email is an error, you should hunt down the
    party responsible for the hookworm instance registered as a WebHook URL in
    this repo's <a href="{{.RepoUrl}}/settings/hooks">service hook settings</a>.
  </p>

  <p>
    Pretty please submit issues specific to hookworm functionality
    <a href="https://github.com/modcloth-labs/hookworm/issues/new">on github</a>
  </p>

</div>

----==ZOMGBOUNDARAAAYYYYY--
`))
)
