#!/usr/bin/env ruby
require 'hookworm-handlers'

exit Hookworm::RogueCommitHandler.new.run!(ARGV)
