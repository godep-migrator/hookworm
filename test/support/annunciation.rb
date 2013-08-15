require_relative 'colors'

module Annunciation
  def announce!(something)
    $stderr.puts "#{MAGENTA}runtests#{RESET}: " <<
                 "#{YELLOW}#{something}#{RESET}"
  end
end
