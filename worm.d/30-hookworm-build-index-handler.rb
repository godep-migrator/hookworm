#!/usr/bin/env ruby
# -*- coding: utf-8 -*-
# vim:fileencoding=utf-8
#+ #### Hookworm Build Index Handler
#+

require_relative '.hookworm_base'

class HookwormBuildIndexHandler
  include HookwormBase

  private

  def handle_travis
    78
  end
end

exit HookwormBuildIndexHandler.new.run!(ARGV) if $PROGRAM_NAME == __FILE__
