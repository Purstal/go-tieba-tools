require 'json'

require './utils.rb'

def main
	obj = JSON.parse(File.read('a.result.json').force_encoding("utf-8"))
	
	records = {}
	
	obj['记录'].each { |record|
		if record['标题'] =~ /(象牙塔|万象)/ && record['贴吧'] =~ /minecraft/ #因为收据数据时犯了错误导致贴吧这一项把时间也接在后面了...
			match_data = /\/p\/([0-9]*)/.match(record['url'])
			tid = match_data[1]
			#p match_data[1]
			
			if records.has_key?(tid)
				if !record['回复']
					records[tid]['楼主'] = true
				end
				time = parse_time(record["时间"])
				if records[tid]['最早时间'] > time
					records[tid]['最早时间'] = time
				end
				next
			end
			
			records[tid] = {
				"标题" => record["标题"],
				"tid" => tid,
				"最早时间" => parse_time(record["时间"]),
				"楼主" => !record["回复"]
			}
		end
	}
	
	File.open('result.json','w') do |f|
		f.puts JSON.generate(records)
	end
	
	require 'CSV'
	
	CSV.open('result.csv','wb') do |csv|
		csv << ["标题","tid","最早时间","楼主"]
		records.each {|tid,r| csv << [r["标题"],tid,r["最早时间"],r["楼主"]]}
	end
end

main()