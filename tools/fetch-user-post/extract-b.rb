require "./utils.rb"

def main
	require 'json'
	obj = JSON.parse(File.read('iamunknown.result.json').force_encoding("utf-8"))

	records = {} # :time => post counts: map[string][int]
	forum_set = {} # :string => :bool

	obj['记录'].each do |r|
		time = parse_time(r['时间'])
		unix_day = time.to_i / 86400 * 86400
		if !records.has_key?(unix_day)
			records[unix_day] = {'#全部#' => 0}
		end

		md = /(.*)\d{4}-\d{2}-\d{2} \d{2}:\d{2}/.match(r['贴吧'])
		forum = md[1]

		if !records[unix_day].has_key?(forum)
			records[unix_day][forum] = 1
		else
			records[unix_day][forum] += 1
		end

		records[unix_day]['#全部#'] += 1
		forum_set[forum] = true
	end

	require 'CSV'

	CSV.open('result-b.csv','wb') do |csv|
		head = ['日期','全部']
		forum_set.each do |k,v|
			head << k
		end
		csv << head

		records.each do |time,counts|
			row = [Time.at(time).strftime("%Y-%m-%d"),counts['#全部#']]
			forum_set.each do |k,v|
				row << counts[k]
			end
			csv << row
		end
	end

end

main()