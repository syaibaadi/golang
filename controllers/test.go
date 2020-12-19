package controllers

import (
	"github.com/labstack/echo/v4"
)

func getHtml(room string) string {
	html := `
	<head>
		<style type="text/css">
			.content-output {
				border:solid 1px;
				padding:5px;
			}
			.btn-stop {
				padding: 5px; 
				width: 120px;
				margin-left:10px;
			}
		</style>
	</head>
	<body>
	<div style="padding: 5px;">
		<input type="text" id="token" style="padding: 5px; width: 400px;" placeholder="access token" title="Access Token" value="` + room + `"/>
	</div>
	<div style="padding: 5px;">
		<input type="text" id="slug" style="padding: 5px; width: 400px;" placeholder="slug" title="Subdomain" value="` + room + `"/>
	</div>
	<div style="padding: 5px;">
		<textarea id="additional_params" style="padding: 5px; width: 400px;height:200px;" placeholder='{"start_date":"1970-01-01","end_date":"1970-01-01"}'></textarea>
	</div>
	<div style="padding: 5px; display:inline-block;">
		<button id="btnconnect" onclick="connect()" style="padding: 5px; width: 120px;">Connect</button>
		<button id="btnstart" onclick="start()" style="padding: 5px; width: 120px;" disabled="disabled">Start</button>
	</div>
	<div style="padding: 5px;">
		<pre id="output"></pre>
	</div>
	<script>
		if (window["WebSocket"]) {
			var xhttp = new XMLHttpRequest();
  			const protocol = document.location.protocol=="https:" ? "wss:" : "ws:";
			var socket = {}; // = new WebSocket(protocol + "//" + document.location.host + "/connect/");

			var token = document.getElementById("token");
			var slug = document.getElementById("slug");
			var job = document.getElementById("job");
			var adparams = document.getElementById("additional_params");
			var output = document.getElementById("output");
			var btnstart = document.getElementById("btnstart");
			var btnconnect = document.getElementById("btnconnect");
			var jobId;
			
			function isJSON(str) {
				if ( /^\s*$/.test(str) ) return false;
				str = str.replace(/\\(?:["\\\/bfnrt]|u[0-9a-fA-F]{4})/g, '@');
				str = str.replace(/"[^"\\\n\r]*"|true|false|null|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?/g, ']');
				str = str.replace(/(?:^|:|,)(?:\s*\[)+/g, '');
				return (/^[\],:{}\s]*$/).test(str);
			}

			function connect() {
				if(slug.value != "") {
					socket = new WebSocket(protocol + "//" + document.location.host + "/connect/"+slug.value);

					socket.onopen = function () {
						output.innerHTML = "<b>Connected.</b>";
						btnstart.disabled = false;
						btnconnect.disabled = true;
						getsample(job.value);
					};
		
					socket.onclose = function () {
						output.innerHTML = "<b>Connection closed.</b>";
						btnstart.disabled = true;
						btnconnect.disabled = false;
					};
		
					socket.onmessage = function (e) {
						res = JSON.parse(e.data);
						if(res.error !== undefined || res.status == "finish" ||  res.status == "error" ||  res.status == "stop") {
							btnstart.disabled = false;
						}
						var divId = "default";
						if(res.job_id != undefined) {							
							jobId = res.job_id;
							divId = jobId;
						}
						var content = document.getElementById(divId);
						var contentres = document.getElementById("res-"+divId);
						if(content != undefined) {
							contentres.innerHTML = JSON.stringify(res, null, 2);
						} else {
							content = document.createElement("div");
							content.setAttribute("id", divId);
							contentres = document.createElement("div");
							contentres.setAttribute("id", "res-"+divId);
							contentres.innerHTML = JSON.stringify(res, null, 2);
							content.setAttribute("class", "content-output");

							content.prepend(contentres);
							b = document.createElement("b");
							b.innerHTML = "Job Id : "+divId;
							br = document.createElement("br");
							content.prepend(br);
							if(divId != "default") {
								btn = document.createElement("button");
								btn.setAttribute("class", "btn-stop");
								btn.setAttribute("onClick", 'stop(this,"'+divId+'");');
								btn.innerHTML = "stop";
								content.prepend(btn);
							}
							content.prepend(b);

							output.append(content);
						}
					};
				}
			}

			function send(action,id) {
				var params = {};
				var jid = "";
				if(isJSON(adparams.value)) {
					params = JSON.parse(adparams.value);
				}
				if(action=="stop") {
					jid = id;
				}
				var merge = Object.assign({}, {action: action,job_id:jid},params);
				socket.send(JSON.stringify(merge));
			}

			function start() {
				send("start");
			}

			function stop(btn,id) {
				send("stop",id);
				btn.disabled = true;
			}
			
			function getsample(id) {
				var url = document.location.protocol + "//" + document.location.host + "/ws_sample_request/" + id;
				xhttp.onreadystatechange = function() {
					if (this.readyState == 4 && this.status == 200) {
						adparams.value = JSON.stringify(JSON.parse(this.responseText), undefined, 2);
					}
				};
				xhttp.open("POST", url, true);
				xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
  				xhttp.send("token="+token.value+"&slug="+slug.value);
			}
		} else {
			output.innerHTML = "<b>Your browser does not support WebSockets.</b>";
		}
	</script>
	</body>`
	return html
}

func Test(c echo.Context) error {
	return c.HTML(200, getHtml(c.Param("token")))
}

func Ping(c echo.Context) error {
	return c.HTML(200, "Welcome")
}
