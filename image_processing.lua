package.cpath = "/usr/local/openresty/lualib/resty/?.so;"
package.cpath = "/usr/local/openresty/lualib/?.so;"
local redis = require "resty.redis"
local red = redis.new()
-- global timeout must bigger than blpop timeout
red:set_timeout(60000)
local ok, err = red.connect(red, '127.0.0.1', '6379')
if not ok then
    ngx.log(ngx.ERR, "failed to connect: ", err)
    ngx.exit(500)
end

local uuid = require "resty.uuid"
local id = uuid.gen20()
local json = require("cjson")
local uri = ngx.var.request_uri
local data = {}
data.uuid = id
data.url = uri
local t = json.encode(data)
ngx.log(ngx.NOTICE, "uri =  ", uri)
if ngx.var.debug ~= "on" then
	local res, err = red:exists(uri)
	if res == 1 then
	    res, err = red:get(uri)
	    if err then
	        ngx.exit(err)
	    else
	        ngx.print(res)
	        ngx.exit(200)
	    end
	end
end

local res, err = red:lpush("taskQueue",t)
if err then
    ngx.log(ngx.ERR, "failed to lpush:", err)
    ngx.exit(500)
end

res, err = red:blpop(id, 30)
if err then
    ngx.log(ngx.ERR, "failed to blpop:", err)
    ngx.exit(500)
end

if res == ngx.null then
    ngx.log(ngx.ERR, "no element popped:", err)
    ngx.status = 408
    ngx.exit(408)
end

if res[2] ~= "200" then
    ngx.status = res[2]
    ngx.exit(res[2])
end

res, err = red:exists(uri)
if res ~= 1 then
    ngx.log(ngx.ERR, "error happens:", err)
    ngx.exit(500)
end

res, err = red:get(uri)
if err then
    ngx.exit(err)
end
ngx.print(res)
 	
local ok, err = red:set_keepalive(10000, 100)
if not ok then
    ngx.log(ngx.ERR, "failed to set keepalive: ", err)
    return
end
ngx.exit(200)
