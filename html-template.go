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
        function toggleButton(element, enable) {
            element.disabled = !enable;
        }

        window.addEventListener("load", function (evt) {
            var elOpen = document.getElementById("open");
            var elClose = document.getElementById("close");
            var elOutput = document.getElementById("output");
            var elMessageCount = document.getElementById("messageCount");
            var elMpS = document.getElementById("mps");
            var messageCount = 0;
            var message = '';
            var intervalMsgFreezed = false;
            // mps stats
            var intervalTick = 500;
            var intervalSize = Math.floor(5000 / intervalTick) + 1;
            var intervalCount = Array(intervalSize).fill(0);
            var intervalDate = Array(intervalSize).fill(0);
            var intervalIndex = 0;
            var ws;

            setInterval(function() {
                elMessageCount.innerText = messageCount;
                var intervalTotal = intervalCount.reduce((a,b) => a + b, 0);
                var intervalIndex2 = (intervalIndex + 1) % intervalSize;
                elMpS.innerText = Math.round(1000 * intervalTotal / (Date.now() - intervalDate[intervalIndex2]));
                var msg;
                try { msg = JSON.stringify(JSON.parse(message), null, 4); }
                catch (e) { msg = message; }
                if (!intervalMsgFreezed) {
                    elOutput.value = msg;
                }
                intervalDate[intervalIndex] = Date.now();
                intervalIndex = intervalIndex2;
                intervalCount[intervalIndex] = 0;
            }, intervalTick);

            toggleButton(elClose, false);
            document.getElementById("open").onclick = function (evt) {
                if (ws) {
                    return false;
                }
                toggleButton(elOpen, false);
                ws = new WebSocket("{{.}}");
                ws.onopen = function (evt) {
                    toggleButton(elOpen, false);
                    toggleButton(elClose, true);
                }
                ws.onclose = function (evt) {
                    ws = null;
                    toggleButton(elOpen, true);
                    toggleButton(elClose, false);
                }
                ws.onmessage = function (evt) {
                    messageCount++;
                    intervalCount[intervalIndex]++;
                    message = evt.data;
                }
                ws.onerror = function (evt) {
                    message = "ERROR: " + (evt.data || "unknown");
                }
                return false;
            };
            document.getElementById("close").onclick = function (evt) {
                if (!ws) {
                    return false;
                }
                toggleButton(elClose, false);
                ws.close();
                return false;
            };
            document.getElementById("freeze").onclick = function (evt) {
                intervalMsgFreezed = !intervalMsgFreezed;
                document.getElementById("freeze").innerText = (intervalMsgFreezed) ? "Unfreeze" : "Freeze";
                return false;
            };
        });
    </script>
</head>

<body style="height:95%">
    <table>
        <tr>
            <td valign="top" width="50%">
                <form>
                    <button id="open">Open</button>
                    <button id="close">Close</button>
                    <button id="freeze">Freeze</button>
                </form>
            </td>
            <td valign="top" width="50%">
            </td>
        </tr>
    </table>
    <div>
        Num of messages <span id="messageCount">0</span> - <span id="mps">0</span> [mps]
    </div>
    <textarea id="output" style="box-sizing: border-box;width:100%;height:95%"></textarea>
</body>

</html>`))

	// html, err := ioutil.ReadFile("test.html")
	// if err != nil {
	// 	log.Printf("Error while reading config.yaml file: \n%v ", err)
	// }
	// homeTemplate = template.Must(template.New("").Parse(string(html)))
}
