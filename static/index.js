window.onload = async () => {
    hljs.initHighlightingOnLoad();
    let socket = new WebSocket(`ws://${location.host}/ws`);
    socket.onmessage = (event) => {
        let msg = JSON.parse(event.data);
        for (const propertyName in msg) {
            document.title = propertyName;
            document.getElementById("content").innerHTML = msg[propertyName];
        }
        let elem = [...document.getElementsByClassName("highlight")];
        for (const i in elem) {
            hljs.highlightBlock(elem[i]);
        }
    };
}
