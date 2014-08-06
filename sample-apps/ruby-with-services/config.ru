require 'sinatra'
require './cloudcredo'
map '/' do
  run CloudCredo
end
