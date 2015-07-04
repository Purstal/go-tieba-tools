def parse_time(str)
	md = /(\d*)-(\d*)-(\d*) (\d*):(\d*)/.match(str)
	year = md[1].to_i; month = md[2].to_i; day = md[3].to_i
	hour = md[4].to_i; minute = md[5].to_i
	
	return Time.new(year, month, day, hour ,minute)
end