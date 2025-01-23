/** @type {WebSocket} */
var socket;

/** pointer_id: button_number */
const holding_buttons = {};

/**
 * @param {HTMLButtonElement} button 
 * @param {number} number 
 */
function pressed(button, number)
{
    if(button.classList == "") return;
    socket.send(number);
}

function get_button_class(button_status)
{
    if(button_status == "0") return "off";
    if(button_status == "1") return "on";
    if(button_status == "-") return "";

    throw new Error("no such character: " + button_status);
}

window.onblur = (e) => {
    for(var k in holding_buttons)
    {
        window.onpointerup({pointerId: k})
    }
}

window.onpointerup = window.onpointercancel = (ev) => {
    const radio_number = holding_buttons[ev.pointerId];
    if(!radio_number) return;

    pressed(document.getElementById(`radio_${radio_number}`), radio_number);
    delete holding_buttons[ev.pointerId];
}

window.onload = () => {
    /** @type {HTMLUListElement} */
    const user_list = document.getElementById("users");

    /** @type {HTMLSpanElement} */
    const user_count_span = document.getElementById("user_count");

    /** @type {HTMLButtonElement[]} */
    const buttons = document.querySelectorAll("#buttons button");

    for(const button of buttons)
    {
        if(button.getAttribute("push") == null) continue;

        button.onpointerdown = (e) => {
            if(e.button != 0) return;

            const number = button.getAttribute("pin_num");
            if(button.querySelector("p") !== null) return;
            pressed(button, number);
            holding_buttons[e.pointerId] = number;
        }
    }

    /** @type {HTMLCanvasElement} */
    const canvas = document.getElementById("video");
    var ctx;
    if(canvas) ctx = canvas.getContext("2d");
    
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
            const blob = new Blob([data], { type: 'image/jpeg' });
            const img = new Image();
            img.onload = () => {
                if(canvas.hidden) canvas.hidden = false;
                
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
            const user_text = data.slice(1);
            const users = user_text.split(",");
            users.pop(); // remove last empty entry
            user_count_span.innerText = users.length;
            user_list.innerHTML = "";
            for(var user of users)
            {
                const li = document.createElement("li");
                li.innerText = user;
                user_list.appendChild(li);
            }
            return;
        }
        else if(data[0] == "h")
        {
            const user_text = data.slice(1);
            const users = user_text.split(",");
            users.pop(); // remove last empty entry
            const user_button_pairs = [];

            for(const user of users)
            {
                const tmp = user.split(";")
                user_button_pairs[tmp[1]] = tmp[0]
            }

            for(const button of buttons)
            {
                for(const p of button.querySelectorAll("p"))
                {
                    button.removeChild(p);
                }

                const user = user_button_pairs[button.getAttribute("pin_num")];
                if(user !== undefined)
                {
                    const p = document.createElement("p");
                    button.appendChild(p);
                    p.innerText = user;
                }
            }

            return;
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
        alert("Connection closed. Reloading webpage.");
        window.location.href = window.location.href;
    };
}
