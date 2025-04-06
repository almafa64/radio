const holding_command = "h";
const user_list_command = "u";
const button_list_command = "b";
const editor_command = "e";

/** @type {WebSocket} */
var socket;

/** pointer_id: button_number */
const holding_buttons = {};

/** @type {HTMLUListElement} */
var user_list;
/** @type {HTMLSpanElement} */
var user_count_span;
/** @type {HTMLButtonElement[]} */
var buttons;

var my_name = "";

/**
 * @param {HTMLButtonElement} button
 * @param {number} number
 */
function pressed(button, number)
{
    if(button.classList == "") return;
    socket.send(number);
}

function users_popup() {
    if(user_list.hidden) {
        user_list.hidden = false;
        return;
    }
    user_list.hidden = true;
}

/**
 * @param {string} button_status
 */
function get_button_class(button_status)
{
    if(button_status == "0") return "off";
    if(button_status == "1") return "on";
    if(button_status == "-") return "";

    throw new Error("no such character: " + button_status);
}

// -------- Events --------

/**
 * @param {string} data
 */
function parse_data(data) {
    const text = data.slice(1);
    const values = text.split(",");
    if(values[values.length - 1] == '') values.pop(); // remove last empty entry
    return values
}

/**
 * @param {string[]} users
 */
function users_change_event(users) {
    user_count_span.innerText = users.length;
    user_list.innerHTML = "";
    for(const user of users)
    {
        const li = document.createElement("li");
        li.innerText = user;
        user_list.appendChild(li);
    }
}

/**
 * @param {string[]} buttons name;number;type(0: push, 1: toggle),...
 */
function buttons_change_event(buttons) {
    const button_holder = document.getElementById("buttons");

    button_holder.innerHTML = "";

    for(const button of buttons) {
        const button_data = button.split(";")
        button_holder.appendChild(create_button(button_data[0], button_data[1], button_data[2] == "1"));
    }

    init_buttons();
}

/**
 * @param {string} name
 */
function editor_user_change_event(name) {
    /** @type {HTMLButtonElement} */
    const editor_but = document.getElementById("editor_but");

    editor_but.disabled = false;
    document.getElementById("current_editor_span").innerText = name || "";

    if(name == undefined)
        exit_editor();
    else if(name == my_name)
        enter_editor();
    else
        editor_but.disabled = true;
}

/**
 * @param {string[]} users
 */
function holding_change_event(users) {
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
}

/**
 * @param {string} data
 */
function pin_status_change_event(data) {
    for(var i = 0; i < data.length; i++)
    {
        buttons[i].classList = get_button_class(data[i]);
    }
}

// -------- Main part --------

function init_buttons() {
    buttons = document.querySelectorAll("#buttons button");

    for(const button of buttons)
    {
        if(button.getAttribute("toggle") != null)
        {
            button.onpointerdown = (e) => {
                if(e.button != 0) return;

                const number = button.getAttribute("pin_num");
                pressed(button, number);
            }
            continue;
        }

        button.onpointerdown = (e) => {
            if(e.button != 0) return;

            if(button.querySelector("p") !== null) return;
            const number = button.getAttribute("pin_num");
            pressed(button, number);
            holding_buttons[e.pointerId] = number;
        }
    }
}

/**
 * @param {string} name 
 * @param {number} num 
 * @param {boolean|number} isToggle 
 * @returns {HTMLButtonElement}
 */
function create_button(name, num, isToggle) {
    const button = document.createElement("button");
    button.setAttribute("pin_num", num);
    if(isToggle) 
        button.setAttribute("toggle", "");
    else
        button.setAttribute("push", "");
    button.innerText = name;
    button.id = `radio_${num}`;
    return button;
}

// When page goes out of focus, depress all held button
window.onblur = (e) => {
    for(const k in holding_buttons)
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
    user_list = document.getElementById("users");
    user_count_span = document.getElementById("user_count");

    const cameras = {};
    var can_receive_frame = true;

    socket = new WebSocket("ws://" + location.host + "/radio_ws");
    socket.binaryType = 'arraybuffer';

    socket.onopen = (event) => {
        console.log("Connected to WebSocket server.");
    };

    socket.onmessage = (event) => {
        const data = event.data;

        if(data instanceof ArrayBuffer) {
            if(!can_receive_frame) return;
            can_receive_frame = false;

            let view = new DataView(data);
            let id = view.getUint8(0);

            if (!(id in cameras)) {
                cameras[id] = document.getElementById(`video${id}`).getContext('2d');
            }

            let ctx = cameras[id];
            let canvas = ctx.canvas;

            const blob = new Blob([data.slice(1)], { type: 'image/jpeg' });

            createImageBitmap(blob)
                .then(img => {
                    if(canvas.hidden)
                    {
                        canvas.hidden = false;
                        canvas.width = img.width;
                        canvas.height = img.height;
                    }
                    ctx.drawImage(img, 0, 0);
                    can_receive_frame = true;
                })
                .catch(err => {
                    console.err("failed to decode frame: ", err);
                    can_receive_frame = true;
                });
            return;
        }

        console.log("Message from server:", data);

        const command = data[0];
        const args = parse_data(data);

        switch (command) {
            case user_list_command:
                if(args[0][0] == '*')
                    my_name = args[0].substring(1)
                else
                    users_change_event(args);
                return;
            case holding_command:
                holding_change_event(args);
                return;
            case button_list_command:
                buttons_change_event(args);
                return;
            case editor_command:
                editor_user_change_event(args[0]);
                return;
        }

        if(buttons.length !== data.length)
        {
            console.log("wrong length of data");
            return;
        }

        pin_status_change_event(data);
    };

    socket.onclose = (event) => {
        alert("Connection closed. Reloading webpage.");
        window.location.href = window.location.href;
    };
}

// -------- Editor --------

function save_pins() {

}

var in_editor = false;

/** @type {HTMLButtonElement[]} */
var editor_buttons = [];

function request_editor() {
    socket.send(editor_command); // send editor access request to server
}

function enter_editor() {
    if(in_editor) return;
    in_editor = true;

    for(const button of buttons)
    {
        /** @type {HTMLButtonElement} */
        const editor_but = button.cloneNode(false);
        button.hidden = true;
        editor_buttons.push(editor_but);
        button.parentElement.appendChild(editor_but);

        editor_but.onclick = () => {
            // open settings popup
        }
    }
}

function exit_editor() {
    if(!in_editor) return;
    in_editor = false;

    for(const button of editor_buttons)
    {
        button.remove();
    }
    for(const button of buttons)
    {
        button.hidden = false;
    }
    editor_buttons = []
}