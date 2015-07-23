function intToBytes(t) {
    var e = t.toString(16),
        n = e.length % 4;
    if (0 != n) for (var a = n; 4 > a; a++) e = "0" + e;
    for (var r = "",
        a = 0; a < e.length; a += 4) {
        var s = (parseInt(e[a], 16) << 12) + (parseInt(e[a + 1], 16) << 8) + (parseInt(e[a + 2], 16) << 4) + parseInt(e[a + 3], 16),
            o = String.fromCharCode(s);
        r += o
    }
    return r
}
function bytesToInt(t) {
    for (var e = 0,
        n = 0; n < t.length; n++) e += t.charCodeAt(n) * Math.pow(65536, t.length - n - 1);
    return e
}
function encodeIntsToString(t) {
    if (0 === t.length) return "";
    for (var e = "",
        n = [intToBytes(t[0])], a = 1; ; a++) {
        var r = !1;
        if (a < t.length) {
            var s = intToBytes(t[a]);
            n[0].length === s.length && (n.push(s), r = !0)
        }
        if (a == t.length || !r) {
            var o = n.length,
                i = intToBytes(o);
            e += String.fromCharCode(n[0].length) + String.fromCharCode(i.length) + i;
            for (var g = 0; g < n.length; g++) e += n[g];
            if (!(a < t.length)) break;
            n = [intToBytes(t[a])]
        }
    }
    return e
}
function decodeIntsFromString(t) {
    for (var e = [], n = 0; n < t.length;) {
        var a = t.charCodeAt(n),
            r = t.charCodeAt(n + 1),
            s = bytesToInt(t.substr(n + 2, r));
        n = n + 2 + r;
        for (var o = 0; s > o; o++) e.push(bytesToInt(t.substr(n, a))),
            n += a
    }
    return e
}
function collectThreadLists() {
    var t = {
        Biggest: "",
        Others: []
    },
        e = 0,
        n = $(".thread-list");
    return n.each(function () {
        var n = $(this).find(".thread").length;
        n > e && (t.Biggest = this.id, e = n)
    }),
        n.each(function () {
            if (this.id === t.Biggest) return !0;
            var e = [],
                n = $(this).find(".thread");
            0 !== n.length && (n.each(function () {
                e.push(parseInt(this.id.substr(1)))
            }), t.Others.push({
                ID: this.id,
                Tids: encodeIntsToString(e)
            }))
        }),
        t
}
function collectData() {
    return {
        PageID: pageID,
        TimeRangeString: $("#subheading").text(),
        ThreadLists: collectThreadLists()
    }
}
function collectThreads() {
    var t = [];
    return $(".thread").each(function () {
        t.push(parseInt(this.id.substr(1)))
    }),
        t
}
function loadData(t) {
    for (var e = collectThreads().sort(function (t, e) {
        return t - e
    }), n = {},
        a = 0; a < e.length; a++) n[e[a]] = !0;
    for (var a = 0; a < t.ThreadLists.Others.length; a++) for (var r = t.ThreadLists.Others[a], s = r.ID, o = decodeIntsFromString(r.Tids), i = 0; i < o.length; i++) {
        var g = o[i];
        n[g] = !1;
        var d = document.getElementById("t" + g);
        "unpassed" === s ? $(d).find(".unpassed-button").attr("style", "display : none") : $(d).find(".unpassed-button").attr("style", "display : inline"),
        $(document.getElementById(s)).append(d)
    }
    for (var g in n) if (n[g]) {
        var l = document.getElementById("t" + g);
        "unpassed" !== t.ThreadLists.Biggest && $(l).find(".unpassed-button").attr("style", "display : inline"),
        $(document.getElementById(t.ThreadLists.Biggest)).append(l)
    }
}
function setJsonData(t) {
    if (t.value === "") {
        t.value = "请把JSON复制到此框内~"
        return
    }
    var e = JSON.parse(t.value);
    e.PageID !== pageID ? showMsg(document.getElementById("text-storage-message"), "页面ID不匹配.") : (loadData(e), showMsg(document.getElementById("text-storage-message"), "载入完成."))
}
function getJsonData(t) {
    t.hidden = !1,
    t.textContent = JSON.stringify(collectData()),
    showMsg(document.getElementById("text-storage-message"), "获取完成.")
}
function saveLocalData() {
    return void 0 === window.localStorage ? void showMsg(document.getElementById("local-storage-message"), "好像不支持localStorage哦>_<.") : (window.localStorage[baseName] = JSON.stringify(collectData()), void showMsg(document.getElementById("local-storage-message"), "保存完成."))
}
function loadLocalData() {
    return void 0 === window.localStorage ? void showMsg(document.getElementById("local-storage-message"), "好像不支持localStorage哦>_<.") : (loadData(JSON.parse(window.localStorage[baseName])), void showMsg(document.getElementById("local-storage-message"), "读取完成."))
}