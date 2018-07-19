window.onload = async () => {
    let socket = new WebSocket(`ws://${location.host}/ws`);
    socket.onmessage = (event) => {
        let msg = JSON.parse(event.data);
    };
}
