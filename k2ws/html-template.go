package main

import (
	"html/template"
)

type templateInfo struct {
	TestPath, WSURL string
}

var homeTemplate *template.Template

func readTemplate() {
	homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html style="height:100%">

<head>
    <meta charset="utf-8">
    <link rel="stylesheet" href="{{.TestPath}}/static/semantic-ui/2.3.1/semantic.min.css" crossorigin="anonymous">
    <link rel="stylesheet" href="{{.TestPath}}/static/semantic-ui/2.3.1/components/dropdown.min.css" crossorigin="anonymous">
    <script src="{{.TestPath}}/static/jquery-3.3.1.slim.min.js" crossorigin="anonymous"></script>
    <script src="{{.TestPath}}/static/semantic-ui/2.3.1/semantic.min.js" crossorigin="anonymous"></script>
    <script src="{{.TestPath}}/static/semantic-ui/2.3.1/components/dropdown.min.js" crossorigin="anonymous"></script>
    <script src="{{.TestPath}}/static/codemirror/5.38.0/codemirror.min.js"></script>
    <link rel="stylesheet" href="{{.TestPath}}/static/codemirror/5.38.0/codemirror.min.css" crossorigin="anonymous">
    <style>
        .CodeMirror { height: 200px; }
    </style>
    <script src="{{.TestPath}}/static/codemirror/5.38.0/mode/javascript/javascript.min.js"></script>
    <script>
        function $f_toggleButton(element, enable) {
            element.disabled = !enable;
        }

        function $f_wsURL() {
            var topics = document.getElementById("topics").value;
            var group_id = document.getElementById("group.id").value;
            var auto_offset = document.getElementById("auto.offset.reset").value;
            return "{{.WSURL}}?topics=" + topics + "&group.id=" + group_id + "&auto.offset.reset=" + auto_offset;
        }

        window.addEventListener("load", function (evt) {
            $('.ui.dropdown').dropdown();
            var $$el = {
                open: document.getElementById("open"),
                close: document.getElementById("close"),
                output: document.getElementById("output"),
                messageCount: document.getElementById("messageCount"),
                mps: document.getElementById("mps"),
                filterBox: document.getElementById("filterbox"),
                filterEditor: undefined,//ace.edit("editor"),
                stacked: document.getElementById("stacked"),
                stopAfterCount: document.getElementById("stopaftercount"),
            }
            $$el.filterEditor = CodeMirror(function(elt) {
                    var e = document.getElementById("editor");
                    e.parentNode.replaceChild(elt, e);
                }, {
                    value: document.getElementById("editor").value,
                    mode:  "javascript"
                });
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

            function $$evalFilter(msg) {
                if ($$flags.filterEnabled) {
                    try {
                        msg = eval('var m=' + msg + ';\n' + $$el.filterEditor.getValue());
                        if (msg && typeof (msg) === 'object') {
                            msg = JSON.stringify(msg, null, 4);
                        }
                    }
                    catch (e) {
                        msg = e.message;
                    }
                }
                return msg;
            }

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
                    msg = $$evalFilter(msg);
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
            document.getElementById("settings").onclick = function (evt) {
                document.getElementById('kafka').style.display = (document.getElementById('kafka').style.display == 'none') ? 'inline' : 'none';
                window.dispatchEvent(new Event('resize'));
                return false;
            }
            document.getElementById("freeze").onchange = function (evt) {
                $$flags.intervalMsgFreezed = !!this.checked;
                this.innerText = ($$flags.intervalMsgFreezed) ? "Unfreeze" : "Freeze";
                return false;
            };
            document.getElementById("filter").onchange = function (evt) {
                $$flags.filterEnabled = !!this.checked;
                $$el.filterBox.style.display = $$flags.filterEnabled ? "block" : "none";
                $$flags.message = '';
                window.dispatchEvent(new Event('resize'));
                return false;
            };
            document.getElementById("eval").onclick = function (evt) {
                $$flags.message = $$evalFilter('""');
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
            document.getElementById("resetcounter").onclick = function() {
                $$flags.messageCount = 0;
            }
            window.dispatchEvent(new Event('resize'));
        });

        window.addEventListener("resize", function (evt) {
            var n = document.getElementById("output").parentNode;
            n.style.height = (window.innerHeight - n.offsetTop - 5) + 'px';
        });
    </script>
</head>

<body style="height:100%; background: #f0f0f0">
    <div>
        <div class="small ui icon buttons">
            <button class="ui button" id="open" title="Open WebSocket">
                <i class="play icon"></i>
            </button>
            <button class="ui button" id="close" title="Stop WebSocket">
                <i class="stop icon"></i>
            </button>
            <button class="ui button" id="settings" title="Setup Kafka Consumer">
                <i class="cog icon"></i>
            </button>
        </div>
        <i class="ellipsis vertical icon"></i>
        <div class="small ui toggle checkbox">
            <input type="checkbox" name="public" id="stopafter">
            <label>Stop after
                <span id="stopafterlimit"></span>
                <span id="stopaftercount"></span>
            </label>
        </div>
        <i class="ellipsis vertical icon"></i>
        <div class="small ui toggle checkbox">
            <input type="checkbox" name="public" id="stack">
            <label>Stack
                <span id="stacked"></span>
            </label>
        </div>
        <i class="ellipsis vertical icon"></i>
        <div class="small ui toggle checkbox">
            <input type="checkbox" name="public" id="filter">
            <label>Filter</label>
        </div>
        <i class="ellipsis vertical icon"></i>
        <div class="small ui toggle checkbox">
            <input type="checkbox" name="public" id="freeze">
            <label>Freeze</label>
        </div>
    </div>
    <div>
        <form>
            <div id="kafka" style="display: none">
                <div class="small ui left icon input">
                    <input id="topics" type="text" placeholder="topic1,topic2,topic3">
                    <i class="plug icon"></i>
                </div>
                <div class="small ui left icon input">
                    <input id="group.id" type="text" placeholder="group id">
                    <i class="users icon"></i>
                </div>
                <select class="small ui dropdown" id="auto.offset.reset">
                    <option value="earliest">earliest</option>
                    <option value="latest" selected="selected">latest</option>
                </select>
            </div>
        </form>
    </div>
    <br/>
    <div id="filterbox" style="display: none; width: 100%;">
        <textarea id="editor">/* "m" is incomming message
 * use "g" as global object if you need
 * last statement becomes final message
 * falsy message will be dropped */
m</textarea>
        <div style="text-align: left">
            <div class="tiny ui icon buttons">
                <button class="ui button" id="eval" title="Evalute">
                    <i class="exclamation icon"></i>
                </button>
            </div>
        </div>
    </div>
    <div style="text-align: right; padding: 2px">
        Num of messages
        <span id="messageCount">0</span> [<span id="mps">0</span>]
        <button class="mini ui button" title="Reset message count" id="resetcounter">
            <i class="redo alternate icon"></i>
        </button>
        <i class="ellipsis vertical icon"></i>
        Refresh rate
        <select class="ui dropdown" id="tick">
            <option value="200">200 ms</option>
            <option value="500">500 ms</option>
            <option value="1000" selected="selected">1 sec</option>
            <option value="5000">5 sec</option>
            <option value="30000">30 sec</option>
        </select>
    </div>
    <div class="field">
        <textarea id="output" style="width:100%;height:100%;"></textarea>
    </div>
</body>

</html>
`))
	// html, err := ioutil.ReadFile("../test.html")
	// if err != nil {
	// 	log.Printf("Error while reading config.yaml file: \n%v ", err)
	// }
	// homeTemplate = template.Must(template.New("").Parse(string(html)))
}
