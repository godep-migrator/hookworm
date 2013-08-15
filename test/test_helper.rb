# vim:fileencoding=utf-8
if ENV['IN_TEST_RUNNER']
  require 'minitest/spec'
else
  require 'minitest/autorun'
end

top = File.expand_path('../../', __FILE__)
unless $LOAD_PATH.include?(top)
  $LOAD_PATH.unshift(top)
end

Dir.glob("#{File.expand_path('../', __FILE__)}/support/*.rb").each do |f|
  require f
end
