width={}
hight={}
for j = 1,10000,1
do
	width[j] = math.random(4096)
    hight[j] = math.random(4096)
end
mode = 2
i=0
--
--request = function(){}
--
function request()
	local x = i
	x = x % #width + 1
	i = i + 1
	local path = "/s3.lecloud.com/test/timg.jpg?imageview/"..mode.."/w/"..width[x].."/h/"..hight[x].."/l/1"
	return wrk.format("GET", path, nil, nil)
end

function response(status, headers, body)  
   if status ~= 200 then  
      print(status)
   end  
end
