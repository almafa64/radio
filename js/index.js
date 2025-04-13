"use strict";

/**
 * @typedef {Object} Module
 * @property {string} Type
 */

/**
 * @typedef {Object} Button
 * @property {string} Name
 * @property {number} Pin
 * @property {number} Default
 * @property {boolean} IsToggle
 */

/**
 * @typedef {Object} ButtonModuleProperties
 * @property {Button[]} Buttons
 * @typedef {Module & ButtonModuleProperties} ButtonModule
 */

/**
 * @typedef {Object} CameraModuleProperties
 * @property {string} Name
 * @property {number} Fps
 * @property {string} Device
 * @property {string} Format
 * @property {string} Resolution
 * @typedef {Module & CameraModuleProperties} CameraModule
 */

/**
 * @typedef {Module[]} Segment
 */

/**
 * @typedef {Object} PageSchemeData
 * @property {Segment[]} Segments
 */

const holding_command = "h";
const user_list_command = "u";
const editor_command = "e";
const json_command = "j";

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

/** @type {Object<number, CanvasRenderingContext2D>} */
const cameras = {};

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
        if(user == my_name) {
            li.style.backgroundColor = "green";
        }
    }
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

        const user = user_button_pairs[button.dataset.pin];
        if(user !== undefined)
        {
            const p = document.createElement("p");
            button.appendChild(p);
            p.innerText = user;
        }
    }
}

// TODO: extend for >9 pin numbers
/**
 * @param {string} data
 */
function pin_status_change_event(data) {
    for(var i = 0; i < data.length; i++)
    {
        buttons[i].classList = get_button_class(data[i]);
    }
}

// -------- JSON events --------

/**
 * @param {ButtonModule} module
 * @param {HTMLDivElement} module_div
 */
function add_button_module(module, module_div) {
    module_div.classList.value = "buttons";
    module_div.innerHTML = "";
    for(const button of module.Buttons) {
        const button_elem = create_button(button.Name, button.Pin, (button.Default == -1) ? 1 : button.IsToggle);
        button_elem.classList.value = get_button_class((button.Default == -1) ? "-" : button.Default);
        module_div.appendChild(button_elem);
        button_elem.dataset.default = button.Default
        button_elem.dataset.isToggle = (button.Default == -1) ? 1 : button.IsToggle
        button_elem.dataset.name = button.Name
        button_elem.dataset.pin = button.Pin
    }
}

/**
 * @param {CameraModule} module 
 * @param {HTMLDivElement} module_div
 */
function add_camera_module(module, module_div, camera_id) {
    module_div.classList.value = "";
    var canvas = module_div.querySelector("canvas");
    if (!canvas) {
        module_div.innerHTML = `<canvas id="video${camera_id}"></canvas><p></p>`;
        canvas = module_div.querySelector("canvas");
        cameras[camera_id] = canvas.getContext("2d");
    }
    module_div.querySelector("p").innerText = module.Name;
    canvas.dataset.format = module.Format
    canvas.dataset.fps = module.Fps
    canvas.dataset.name = module.Name
    canvas.dataset.device = module.Device
    canvas.dataset.resolution = module.Resolution
    canvas.dataset.type = module.Type
}

/**
 * @param {PageSchemeData} data 
 */
function page_scheme_event(data) {
    var camera_counter = 0;

    var remove_segments = [...document.getElementsByClassName("segments")];
    var remove_modules = [...document.querySelectorAll(".segments > div")];

    for(const segment_idx in data.Segments) {
        const segment = data.Segments[segment_idx];
        let id = `segment${segment_idx}`;
        let segment_div = document.getElementById(id);
        if (!segment_div) {
            segment_div = document.createElement("div");
            segment_div.id = id;
            segment_div.classList.value = "segments";
            document.body.appendChild(segment_div);
        } else {
            remove_segments = remove_segments.filter(v => v !== segment_div)
        }
        
        for(const module_idx in segment) {
            const module = segment[module_idx];
            id = `module${segment_idx}-${module_idx}`;
            let module_div = document.getElementById(id);
            if (!module_div) {
                module_div = document.createElement("div");
                module_div.id = id;
                segment_div.appendChild(module_div);
            } else {
                remove_modules = remove_modules.filter(v => v !== module_div)
            }

            switch (module.Type) {
                case "buttons": add_button_module(module, module_div); break;
                case "cam":     add_camera_module(module, module_div, camera_counter); camera_counter++; break;
            }
        }
    }

    for(const segment of remove_segments) {
        segment.remove()
    }

    for(const module of remove_modules) {
        module.remove()
    }

    buttons = document.querySelectorAll(".buttons button");
}

/**
 * @param {string} data 
 */
function json_event(data) {
    data = JSON.parse(data)
    const event_data = data["Data"];
    switch(data["Event"]) {
        case "page_scheme": page_scheme_event(event_data); break;
    }
}

// -------- Main part --------

/**
 * @param {string} name 
 * @param {number} num 
 * @param {boolean|number} isToggle 
 * @returns {HTMLButtonElement}
 */
function create_button(name, num, isToggle) {
    const button = document.createElement("button");
    button.innerText = name;
    button.id = `radio_${num}`;

    if(isToggle)
    {
        button.onpointerdown = (e) => {
            if(e.button != 0) return;

            const number = button.dataset.pin;
            pressed(button, number);
        }
    } else {
        button.onpointerdown = (e) => {
            if(e.button != 0) return;

            if(button.querySelector("p") !== null) return;
            const number = button.dataset.pin;
            pressed(button, number);
            holding_buttons[e.pointerId] = number;
        }
    }

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
                /** @type {HTMLCanvasElement} */
                const canvas = document.getElementById(`video${id}`);
                if (!canvas) return;
                cameras[id] = canvas.getContext('2d');
            }

            let ctx = cameras[id];
            let canvas = ctx.canvas;

            // TODO: use attribute "format"
            const blob = new Blob([data.slice(1)], { type: 'image/jpeg' });

            // TODO: clear image after not reciving frames for x time
            createImageBitmap(blob)
                .then(img => {
                    canvas.width = img.width;
                    canvas.height = img.height;
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
            case editor_command:
                editor_user_change_event(args[0]);
                return;
            case json_command:
                json_event(args);
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