#!/usr/bin/env ruby
require 'hookworm-handlers'

exit Hookworm::BuildIndexHandler.new.run!(ARGV)
