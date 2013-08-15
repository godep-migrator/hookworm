require_relative 'annunciation'

class MiniTestReporter
  include Annunciation

  def puts(*args)
    args.each { |arg| announce! arg }
  end

  alias print puts
end
