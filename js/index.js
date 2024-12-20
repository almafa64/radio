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
    /** @type {HTMLUListElement} */
    const user_list = document.getElementById("users");

    /** @type {HTMLSpanElement} */
    const user_count_span = document.getElementById("user_count");

    /** @type {HTMLButtonElement[]} */
    const buttons = document.querySelectorAll("#buttons button");

    /** @type {HTMLCanvasElement} */
    const canvas = document.getElementById("video");
    const ctx = canvas.getContext("2d");
    
    var can_recive_frame = true;

    socket = new WebSocket("ws://" + location.host + "/radio_ws");
    socket.binaryType = 'arraybuffer';

    socket.onopen = (event) => {
        console.log("Connected to WebSocket server.");
    };

    socket.onmessage = (event) => {
        const data = event.data;

        if(data instanceof ArrayBuffer) {
            if(!can_recive_frame) return;
            
            can_recive_frame = false;
            const blob = new Blob([data], { type: 'video/h264' });
            const img = new Image();
            img.onload = () => {
                ctx.drawImage(img, 0, 0);
                can_recive_frame = true;
                URL.revokeObjectURL(img.src);
            }
            img.onerror = () => {
                console.error("frame dropped");
                can_recive_frame = true;
                URL.revokeObjectURL(img.src);
            };
            img.src = URL.createObjectURL(blob);
            return;
        }

        console.log("Message from server:", data);
        
        if(data === "closed")
        {
            alert("websocket closed")
            return;
        }

        if(data[0] == "u")
        {
            const user_text = data.slice(1)
            const users = user_text.split(",")
            users.pop() // remove last empty entry
            user_count_span.innerText = users.length;
            user_list.innerHTML = "";
            for(var user of users)
            {
                const li = document.createElement("li");
                li.innerText = user;
                user_list.appendChild(li);
            }
            return
        }

        if(buttons.length !== data.length)
        {
            console.log("wrong length of data");
            return;
        }

        for(var i = 0; i < data.length; i++)
        {
            buttons[i].classList = get_button_class(data[i]);
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket connection closed.");
    };
}
