package save_to_html

const MAIN_TEMPLATE_FIRST_HALF = `
<!DOCTYPE html>
<head>
	<meta charset="UTF-8">
	<link rel="stylesheet" href="http://mrcoles.com/media/test/markdown-css/markdown.css" type="text/css" />
	<!--↑然而并没有什么用-->
	<span id="page-id" content="%s">
</head>
<body>
<script src="http://code.jquery.com/jquery-latest.js"></script>
<script>

function moveTo(where, whom) {
	$("#"+where.id).append(whom);
	$("#"+whom.id).find("button").each(function() { /*console.log(this);*/ this.style.display = ""; })
	if (where === undetermined) {
		//console.log($("#"+whom.id).find("button.undetermined_button"))
		$("#"+whom.id).find("button.undetermined_button").attr("style", "display : none");
	} else if (where === passed) {
		$("#"+whom.id).find("button.passed_button").attr("style", "display : none");
	} else if (where === not_passed) {
		$("#"+whom.id).find("button.not_passed_button").attr("style", "display : none");
	}
}

function collectData(whom) {
	var data = []
	$("ul#"+whom).find("span.thread").each(function(i) { data[i] = this.id })
	return data
}

var baseName = "ivory-tower>>"+$("span#page-id").attr("content")

function saveData() {
	localStorage[baseName] = JSON.stringify({
		"unchecked" : collectData("unchecked"),
		"undetermined" : collectData("undetermined"),
		"passed" : collectData("passed"),
		"not_passed" : collectData("not_passed"),
	})

	$("font#msg").attr("style", "display:inline")
	document.getElementById("msg").innerText = "保存完毕"
}

function loadData() {
	var where_ids = ["unchecked","undetermined","passed","not_passed"]
	var data = JSON.parse(localStorage[baseName])
	for (var i = 0; i < where_ids.length ; i++) {
		var where = document.getElementById(where_ids[i])
		var ids = data[where_ids[i]]
		for (var j = 0; j < ids.length; j++ ) {
			moveTo(where, document.getElementById(ids[j]))
		}
	}
	$("font#msg").attr("style", "display:inline")
	document.getElementById("msg").innerText = "读取完毕"
}

/* 但其实不是重置..干脆去掉算了..
function resetAllData() {
	if (!this.confirm) {
		this.confirm = true
		this.innerText = "点击确认重置!"
		$("font#msg").attr("style", "display:inline")
		document.getElementById("msg").innerText = "确认重置请再按一次"
	} else {
		this.confirm = false
		this.innerText = "重置"
		
		var where_ids = ["undetermined","passed","not_passed"]
		for (var i = 0; i < where_ids.length ; i++) {
			resetData(where_ids[i])
		}
		$("font#msg").attr("style", "display:inline")
		document.getElementById("msg").innerText = "重置完毕"
	}
}

function resetData(where_id) {
	$("ul#"+where_id).find("span.thread").each(function(){
		moveTo(unchecked, this)
	})
}
*/

</script>

<h1>%s</h1>

<button onclick="loadData()">读取</button>
<button onclick="saveData()">保存</button>
<!--<button onclick="resetAllData()" id="reset_button">重置</button>-->

<font id="msg" style="display:none"></font>

<h1>未检查</h1>
<ul id="unchecked">`

const MAIN_TEMPLATE_SECOND_HALF = `
</ul>

<h1>待定</h1>
<ul id="undetermined">
</ul>

<h1>通过</h1>
<ul id="passed">
</ul>

<h1>不通过</h1>
<ul id="not_passed">
</ul>

</body>
`

var MAIN_TEMPLATE_SECOND_HALF_bytes = []byte(MAIN_TEMPLATE_SECOND_HALF)

type Record struct {
	Tid      uint64
	Title    string
	Author   string
	Time     string
	Abstract string
}

//tid, title, tid, tid, tid, author, tid, tid, time, abstract
const RECORD_TEMPLATE = `
	<span id="t%d" class="thread">
		<li>
			<h2>%s</h2>
			<button onclick="moveTo(undetermined, t%d)" class="undetermined_button">待定</button>
			<button onclick="moveTo(passed, t%d)" class="passed_button">通过</button>
			<button onclick="moveTo(not_passed, t%d)" class="not_passed_button">不通过</button>
			<ul>
				<li>作者: %s</li>
				<li>tid: <a href="http:tieba.baidu.com/p/%d" target="_Blank">%d</a></li>
				<li>时间:%s</li>
				<li>摘要:<ul>
					%s
				</ul></li>
			</ul>
		</li>
	</span>
`
