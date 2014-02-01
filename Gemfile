source 'https://rubygems.org'

if ENV['DEV_MODE']
  gem 'hookworm-handlers',
      path: File.join(ENV['HOME'], 'workspace/hookworm-handlers-ruby')
else
  gem 'hookworm-handlers', '~> 0.1'
end

group :development do
  gem 'mail', '~> 2.5'
  gem 'rubocop', '~> 0.17'
end
