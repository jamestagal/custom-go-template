let bodyEl = document.getElementsByTagName("body")[0];
let cms_fields_str = bodyEl.getAttribute("x-data");
let cms_fields = eval("(" + cms_fields_str + ")");

function createInputs(obj, container) {
    for (const key in obj) {
        if (Object.hasOwnProperty.call(obj, key)) {
            const input = document.createElement('input');
            input.type = 'text';
            let attribute = "x-model";
            if (typeof obj[key] === 'number') {
                input.type = 'number';
                attribute += ".number";
            }
            input.id = key;
            input.name = key;
            input.placeholder = key;
            input.setAttribute(attribute, key);
            const label = document.createElement('label');
            label.htmlFor = key;
            label.textContent = key;
            const div = document.createElement('div').appendChild(label).parentNode;
            container.appendChild(document.createElement('br'));
            container.appendChild(document.createElement('br'));
            container.appendChild(div);
            container.appendChild(input);
        }
    }
}

// CRITICAL FIX: Add null guard for missing #plenti_cms element
const cms = document.getElementById('plenti_cms');
if (cms) {
    createInputs(cms_fields, cms);
} else {
    console.warn('[CMS] #plenti_cms element not found in DOM - CMS UI disabled');
}

// CRITICAL FIX: Add null guard for toggle button
const toggleButton = document.getElementById('toggle_plenti_cms');
if (toggleButton) {
    toggleButton.addEventListener('click', function () {
        const menu = document.getElementById('plenti_cms');
        if (menu) {
            menu.classList.toggle('menu-visible');
        }
    });
} else {
    console.warn('[CMS] #toggle_plenti_cms button not found - CMS toggle disabled');
}
