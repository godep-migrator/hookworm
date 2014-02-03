#!/usr/bin/env ruby
require 'hookworm-handlers'

exit Hookworm::Annotator.new.run!(ARGV)
