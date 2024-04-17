window.addEventListener("DOMContentLoaded", (_) => { 
	let websocket = new WebSocket("ws://" + window.location.host + "/start-processing");
	let room = document.getElementById("status");
	path = window.location.pathname;
  
	websocket.addEventListener("message", function (e) {
	  let data = e.data;
		let p = document.createElement("p")
		p.innerHTML = `<p><strong>${data}</strong></p>`;
		room.append(p);
		room.scrollTop = room.scrollHeight;
	});
});
