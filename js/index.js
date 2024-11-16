/** @type {WebSocket} */
var socket;

async function pressed(button, number)
{
    if(button.classList == "") return;

    socket.send(number - 1);
}

function get_button_class(button_status)
{
    if(button_status == "0") return "off";
    if(button_status == "1") return "on";
    if(button_status == "-") return "";

    throw new Error("no such character: " + button_status);
}

window.onload = () => {
    const buttons = document.querySelectorAll("#buttons button");

    socket = new WebSocket("ws://" + location.host + "/radio_ws");

    socket.onopen = (event) => {
        console.log("Connected to WebSocket server.");
    };

    socket.onmessage = (event) => {
        const text = event.data;
        console.log("Message from server:", text);
        
        if(text === "closed")
        {
            alert("websocket closed")
            return;
        }

        if(buttons.length !== text.length)
        {
            console.log("wrong length of data");
            return;
        }

        for(var i = 0; i < text.length; i++)
        {
            buttons[i].classList = get_button_class(text[i]);
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket connection closed.");
    };
}
