package main

import (
	"html/template"
)

var homeTemplate *template.Template

func readTemplate() {
	homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html style="height:95%">

<head>
    <meta charset="utf-8">
    <script>
        function $f_toggleButton(element, enable) {
            element.disabled = !enable;
        }

        function $f_wsURL() {
            var topics = document.getElementById("topics").value;
            var group_id = document.getElementById("group.id").value;
            var auto_offset = document.getElementById("auto_offset").value;
            return "{{.}}?topics=" + topics + "&group.id=" + group_id + "&auto_offset=" + auto_offset;
        }

        window.addEventListener("load", function (evt) {
            var $$el = {
                open: document.getElementById("open"),
                close: document.getElementById("close"),
                output: document.getElementById("output"),
                messageCount: document.getElementById("messageCount"),
                mps: document.getElementById("mps"),
                filterBox: document.getElementById("filterbox"),
                stacked: document.getElementById("stacked"),
                stopAfterCount: document.getElementById("stopaftercount"),
            }
            var $$flags = {
                messageCount: 0,
                message: '',
                intervalMsgFreezed: false,
                filterEnabled: false,
                stackingEnabled: false,
                stackedMessages: 0,
                stopAfter: 0,
            }
            var $$stats = {
                tick: 0,
                size: 0,
                count: [],
                date: [],
                index: 0,
            }
            // ws handle
            var $$ws;
            // filter global
            var g = {};

            function $f_setupInterval(tick) {
                $$stats.tick = tick;
                $$stats.size = Math.floor(5000 / $$stats.tick) + 1;
                $$stats.count = Array($$stats.size).fill(0);
                $$stats.date = Array($$stats.size).fill(0);
                $$stats.index = 0;
            }
            function $f_statsTick() {
                $$el.stopAfterCount.innerText = $$flags.stopAfterCount ? $$flags.stopAfterCount : '';
                $$el.messageCount.innerText = $$flags.messageCount;
                var intervalTotal = $$stats.count.reduce((a, b) => a + b, 0);
                var index2 = ($$stats.index + 1) % $$stats.size;
                $$el.mps.innerText = Math.round(1000 * intervalTotal / (Date.now() - $$stats.date[index2]));
                var msg = '';
                if ($$flags.message) {
                    try { msg = JSON.stringify(JSON.parse($$flags.message), null, 4); }
                    catch (e) {
                        try { msg = JSON.stringify(JSON.parse('[' + $$flags.message + ']'), null, 4); }
                        catch (e) { msg = $$flags.message; }
                    }
                }
                if (!$$flags.intervalMsgFreezed) {
                    $$el.output.value = msg;
                    $$el.stacked.innerText = $$flags.stackedMessages ? $$flags.stackedMessages : '';
                }
                $$stats.date[$$stats.index] = Date.now();
                $$stats.index = index2;
                $$stats.count[$$stats.index] = 0;
                setTimeout($f_statsTick, $$stats.tick);
            }
            $f_setupInterval(500);
            setTimeout($f_statsTick, $$stats.tick);

            $f_toggleButton($$el.close, false);
            document.getElementById("open").onclick = function (evt) {
                if ($$ws) {
                    return false;
                }
                $$flags.stopAfterCount = 0;
                $f_toggleButton($$el.open, false);
                $$ws = new WebSocket($f_wsURL());
                $$ws.onopen = function (evt) {
                    $f_toggleButton($$el.open, false);
                    $f_toggleButton($$el.close, true);
                }
                $$ws.onclose = function (evt) {
                    $$ws = null;
                    $f_toggleButton($$el.open, true);
                    $f_toggleButton($$el.close, false);
                }
                $$ws.onmessage = function (evt) {
                    var msg = evt.data;
                    if ($$flags.filterEnabled) {
                        try {
                            msg = eval('var m=' + evt.data + ';\n' + $$el.filterBox.value);
                            if (msg && typeof (msg) === 'object') {
                                msg = JSON.stringify(msg, null, 4);
                            }
                        }
                        catch (e) {
                            msg = e.message;
                        }
                    }
                    $$flags.messageCount++;
                    $$stats.count[$$stats.index]++;
                    if (msg) {
                        if ($$flags.stackingEnabled && !$$flags.intervalMsgFreezed) {
                            var splitter = ($$flags.message) ? ',' : '';
                            $$flags.message += splitter + msg;
                            $$flags.stackedMessages++;
                        } else {
                            $$flags.message = msg;
                        }
                        if ($$flags.stopAfter > 0) {
                            $$flags.stopAfterCount++;
                            if ($$flags.stopAfterCount >= $$flags.stopAfter) {
                                document.getElementById("close").click();
                            }
                        }
                    }
                }
                $$ws.onerror = function (evt) {
                    $$flags.message = "ERROR: " + (evt.data || "unknown");
                }
                return false;
            };
            document.getElementById("close").onclick = function (evt) {
                if (!$$ws) {
                    return false;
                }
                $f_toggleButton($$el.close, false);
                $$ws.close();
                return false;
            };
            document.getElementById("freeze").onchange = function (evt) {
                $$flags.intervalMsgFreezed = !!this.checked;
                this.innerText = ($$flags.intervalMsgFreezed) ? "Unfreeze" : "Freeze";
                return false;
            };
            document.getElementById("filter").onchange = function (evt) {
                $$flags.filterEnabled = !!this.checked;
                $$el.filterBox.style.display = $$flags.filterEnabled ? "block" : "none";
                $$flags.message = '';
                return false;
            };
            document.getElementById("tick").onchange = function (evt) {
                $f_setupInterval(+this.value)
                return false;
            };
            document.getElementById("stack").onchange = function (evt) {
                $$flags.stackingEnabled = !!this.checked
                if ($$flags.stackingEnabled) {
                    $$flags.message = '';
                    $$flags.stackedMessages = 0;
                }
                return false;
            };
            document.getElementById("stopafter").onchange = function (evt) {
                if (!!this.checked) {
                    p = prompt("Stop after this number of messages:", $$flags.stopAfter > 0 ? $$flags.stopAfter : 10);
                    if (p === null || isNaN(+p) || +p <= 0) {
                        this.checked = false;
                    } else {
                        $$flags.stopAfter = +p;
                        $$flags.stopAfterCount = 0;
                    }
                } else {
                    $$flags.stopAfter = 0;
                    $$flags.stopAfterCount = 0;
                }
                document.getElementById("stopafterlimit").innerText = $$flags.stopAfter > 0 ? $$flags.stopAfter + ' :' : '';
                return false;
            };
        });
    </script>
</head>

<body style="height:95%">
    <table width="100%" height="100%">
        <tr>
            <td valign="top">
                <form>
                    <button id="open">Open</button>
                    <button id="close">Close</button>
                    <button onclick="document.getElementById('kafka').style.display = (document.getElementById('kafka').style.display == 'none') ? 'inline' : 'none'; return false">...</button>
                    <span id="kafka" style="display: none">
                        <input id="topics" placeholder="topic1,topic2,topic3" type="text"></input>
                        <input id="group.id" placeholder="group id" type="text"></input>
                        <select id="auto_offset">
                            <option value="earliest">earliest</option>
                            <option value="latest" selected="selected">latest</option>
                        </select>
                    </span>
                </form>
                Num of messages
                <span id="messageCount">0</span> -
                <span id="mps">0</span> [mps]
            </td>
            <td valign="bottom" align="right">
                <table>
                    <tr>
                        <td align="left">
                            <input type="checkbox" id="stopafter">Stop after <span id="stopafterlimit"></span></input>
                            <span id="stopaftercount"></span>
                            <br/>
                            <input type="checkbox" id="stack">Stack</input>
                            <span id="stacked"></span>
                            <br/>
                            <input type="checkbox" id="filter">Filter</input>
                            <br/>
                            <textarea id="filterbox" cols="50" rows="5" style="display: none">// "m" is incomming message
// use "g" as global object if you need
// last statement becomes final message
// falsy message will be filtered
m</textarea>
                            <input type="checkbox" id="freeze">Freeze</input>
                            <br/>
                            <select id="tick">
                                <option value="200">200 ms</option>
                                <option value="500" selected="selected">500 ms</option>
                                <option value="1000">1 sec</option>
                                <option value="5000">5 sec</option>
                                <option value="30000">30 sec</option>
                            </select>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
        <tr height="100%">
            <td colspan="2">
                <textarea id="output" style="box-sizing:border-box;width:100%;height:100%;"></textarea>
            </td>
        </tr>
    </table>
</body>

</html>`))

	// html, err := ioutil.ReadFile("test.html")
	// if err != nil {
	// 	log.Printf("Error while reading config.yaml file: \n%v ", err)
	// }
	// homeTemplate = template.Must(template.New("").Parse(string(html)))
}
