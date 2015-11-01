function showThreads(threadLists, t) {
    for (var a = $("#thread-template").html(), n = Handlebars.compile(a), o = $("#abstract-template").html(), l = Handlebars.compile(o), r = Handlebars.compile($("#option-template").html()), s = "", i = 0; i < t.length; i++) {
        var c = t[i];
        s += r({
            Value: c.Class,
            Name: c.Name
        })
    }
    for (var j = 0; j < threadLists.length; j++){
        var date = threadLists[j].Date
        var ThreadList = threadLists[j].ThreadList
        for (var d = "",
            i = ThreadList.length - 1; i >= 0; i--) {
            var m = ThreadList[i],
                u = "";
            if (null !== m.Abstracts) {
                for (var h = !1,
                    p = 0; p < m.Abstracts.length; p++) {
                    var g = m.Abstracts[p];
                    "图片" == g.Type ? (h || (u += "<li>"), u += "<img src=" + g.Content + ' height=200 class="img-thumbnail abstract-image" />', h = !0) : (h && (u += "</li>", h = !1), u += l("图片+_+" == g.Type ? {
                        Type: "图片",
                        Content: g.Content
                    } : g))
                }
                h && (u += "</li>")
            }
            //var T = new Date(1e3 * m.Time);
            d += n({
                ID: "t" + m.Tid,
                Tid: m.Tid,
                Title: m.Title,
                Author: m.Author,
                Abstract: u,
                Time: date,
                SelectHTML: s
            })
        }
    }

    document.getElementById("unchecked").innerHTML = d
}
function init(e, t) {
    document.title = "象~牙~塔~ " + e.TimeRange,
    document.getElementById("subheading").textContent = e.TimeRange,
    document.getElementById("page-id-p").textContent = e.PageID,
    pageID = e.PageID,
    baseName = "ivory-tower>>" + pageID;
    var a = makeSelectHTML(t.Categorys);
    showThreads(e.ThreadLists, t.Categorys)
    constructCategorys(t.Categorys)
    //constructFloatBar(a)
}
function constructCategorys(e) {
    var t = Handlebars.compile($("#category-template").html());
    $(".has-category").each(function () {
        for (var a = "",
            n = 0; n < e.length; n++) a += t({
                Category: e[n].Name,
                ID: this.id + "::" + e[n].Class,
                Class:  + e[n].Class + '_category'
            });
        this.innerHTML = a
    })
}
function constructFloatBar(e) {
    document.getElementById("jump_select").innerHTML = e,
    window.onscroll = function () {
        var e = document.getElementById("float-bar");
        e.style.top = document.body.scrollTop + "px"
    },
    window.onresize = window.onscroll
}
function makeSelectHTML(e) {
    for (var t = "",
        a = Handlebars.compile($("#option-template").html()), n = ["未检查", "待定", "通过", "不通过"], o = ["unchecked", "undetermined", "passed", "unpassed"], l = 0; l < n.length; l++) {
        var r = n[l],
            s = o[l];
        if ("未检查" == r || "不通过" == r) t += a({
            Value: "header-of-" + s,
            Name: r
        });
        else for (var i = 0; i < e.length; i++) t += a({
            Value: "header-of-" + s + "::" + e[i].Class,
            Name: r + "::" + e[i].Name
        })
    }
    return t
}
function jumpTo(e) {
    document.location.hash = e.options[e.selectedIndex].value
}
function moveThreadTo(e, t) {
    if ("unpassed" === e.id) return $(t).find(".unpassed-button").attr("style", "display : none"),
        void $(e).append(t);
    $(t).find(".unpassed-button").attr("style", "");
    var a = $(t).find(".destination-selector").val();
    $(document.getElementById(e.id + "::" + a)).append(t)
}
function showMsg(e, t) {
    e.style.display = "inline",
    e.textContent = t
}
var pageID, baseName;
